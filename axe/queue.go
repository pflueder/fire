package axe

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gopkg.in/tomb.v2"

	"github.com/256dpi/fire"
	"github.com/256dpi/fire/coal"
)

// Blueprint describes a job to enqueued.
type Blueprint struct {
	// The job to be enqueued.
	Job Job

	// The initial delay. If specified the job will not be dequeued until the
	// specified time has passed.
	Delay time.Duration

	// The job period. If given, and a label is present, the job will only
	// enqueued if no job has been finished in the specified duration.
	Period time.Duration
}

type board struct {
	sync.Mutex
	jobs map[coal.ID]*Model
}

// Options defines queue options.
type Options struct {
	// The store used to manage jobs.
	Store *coal.Store

	// The maximum amount of lag that should be applied to every dequeue attempt.
	//
	// By default multiple processes compete with each other when getting jobs
	// from the same queue. An artificial lag limits multiple simultaneous
	// dequeue attempts and allows the worker with the smallest lag to dequeue
	// the job and inform the other workers to prevent parallel dequeue attempts.
	//
	// Default: 100ms.
	MaxLag time.Duration

	// The duration after which a job may be returned again from the board.
	//
	// It may take some time until the board is updated with the new state of
	// the job after a dequeue. The block period prevents another worker from
	// simultaneously trying to dequeue the job. If the initial worker failed to
	// dequeue the job it will be available again after the defined period.
	//
	// Default: 10s.
	BlockPeriod time.Duration

	// The report that is called with job errors.
	Reporter func(error)
}

// Queue manages job queueing.
type Queue struct {
	opts   Options
	tasks  map[string]*Task
	boards map[string]*board
	tomb   tomb.Tomb
}

// NewQueue creates and returns a new queue.
func NewQueue(opts Options) *Queue {
	// set default max lag
	if opts.MaxLag == 0 {
		opts.MaxLag = 100 * time.Millisecond
	}

	// set default block period
	if opts.BlockPeriod == 0 {
		opts.BlockPeriod = 10 * time.Second
	}

	return &Queue{
		opts:  opts,
		tasks: make(map[string]*Task),
	}
}

// Add will add the specified task to the queue.
func (q *Queue) Add(task *Task) {
	// safety check
	if q.boards != nil {
		panic("axe: unable to add task to running queue")
	}

	// prepare task
	task.prepare()

	// get name
	name := GetMeta(task.Job).Name

	// check existence
	if q.tasks[name] != nil {
		panic(fmt.Sprintf(`axe: task with name "%s" already exists`, name))
	}

	// save task
	q.tasks[name] = task
}

// Enqueue will enqueue a job using the specified blueprint.
func (q *Queue) Enqueue(job Job, delay, period time.Duration) (bool, error) {
	return Enqueue(nil, q.opts.Store, job, delay, period)
}

// Callback is a factory to create callbacks that can be used to enqueue jobs
// during request processing.
func (q *Queue) Callback(matcher fire.Matcher, cb func(ctx *fire.Context) Blueprint) *fire.Callback {
	return fire.C("axe/Queue.Callback", matcher, func(ctx *fire.Context) error {
		// get blueprint
		bp := cb(ctx)

		// check if controller uses same store
		if q.opts.Store == ctx.Controller.Store {
			// enqueue job using context store
			_, err := Enqueue(ctx, ctx.Store, bp.Job, bp.Delay, bp.Period)
			if err != nil {
				return err
			}
		} else {
			// enqueue job using queue store
			_, err := q.Enqueue(bp.Job, bp.Delay, bp.Period)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Action is a factory to create an action that can be used to enqueue jobs.
func (q *Queue) Action(methods []string, cb func(ctx *fire.Context) Blueprint) *fire.Action {
	return fire.A("axe/Queue.Callback", methods, 0, func(ctx *fire.Context) error {
		// get blueprint
		bp := cb(ctx)

		// check if controller uses same store
		if q.opts.Store == ctx.Controller.Store {
			// enqueue job using context store
			_, err := Enqueue(ctx, ctx.Store, bp.Job, bp.Delay, bp.Period)
			if err != nil {
				return err
			}
		} else {
			// enqueue job using queue store
			_, err := q.Enqueue(bp.Job, bp.Delay, bp.Period)
			if err != nil {
				return err
			}
		}

		// respond with an empty object
		err := ctx.Respond(fire.Map{})
		if err != nil {
			return err
		}

		return nil
	})
}

// Run will start fetching jobs from the queue and process them. It will return
// a channel that is closed once the queue has been synced and is available.
func (q *Queue) Run() chan struct{} {
	// initialize boards
	q.boards = make(map[string]*board)

	// create boards
	for _, task := range q.tasks {
		name := GetMeta(task.Job).Name
		q.boards[name] = &board{
			jobs: make(map[coal.ID]*Model),
		}
	}

	// prepare channel
	synced := make(chan struct{})

	// run process
	q.tomb.Go(func() error {
		return q.process(synced)
	})

	return synced
}

// Close will close the queue.
func (q *Queue) Close() {
	// kill and wait
	q.tomb.Kill(nil)
	_ = q.tomb.Wait()
}

func (q *Queue) process(synced chan struct{}) error {
	// start tasks
	for _, task := range q.tasks {
		task.start(q)
	}

	// reconcile jobs
	var once sync.Once
	stream := coal.Reconcile(q.opts.Store, &Model{}, func() {
		once.Do(func() { close(synced) })
	}, func(model coal.Model) {
		q.update(model.(*Model))
	}, func(model coal.Model) {
		q.update(model.(*Model))
	}, nil, q.opts.Reporter)

	// await close
	<-q.tomb.Dying()

	// close stream
	stream.Close()

	return tomb.ErrDying
}

func (q *Queue) update(job *Model) {
	// get board
	board, ok := q.boards[job.Name]
	if !ok {
		return
	}

	// lock board
	board.Lock()
	defer board.Unlock()

	// handle job
	switch job.Status {
	case StatusEnqueued, StatusDequeued, StatusFailed:
		// apply random lag
		lag := time.Duration(rand.Int63n(int64(q.opts.MaxLag)))
		job.Available = job.Available.Add(lag)

		// update job
		board.jobs[job.ID()] = job
	case StatusCompleted, StatusCancelled:
		// remove job
		delete(board.jobs, job.ID())
	}
}

func (q *Queue) get(name string) (coal.ID, bool) {
	// get board
	board := q.boards[name]

	// lock board
	board.Lock()
	defer board.Unlock()

	// get time
	now := time.Now()

	// return first available job
	for _, job := range board.jobs {
		if job.Available.Before(now) {
			// block job until the specified timeout has been reached
			job.Available = job.Available.Add(q.opts.BlockPeriod)

			return job.ID(), true
		}
	}

	return coal.ID{}, false
}
