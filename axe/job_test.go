package axe

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/256dpi/fire/stick"
)

type bsonJob struct {
	Base `bson:"-" axe:"bson"`
	Data string `bson:"data"`
}

type invalidJob1 struct {
	Hello string
	Base
}

type invalidJob2 struct {
	Base  `axe:"foo/bar"`
	Hello string
}

type invalidJob3 struct {
	Base  `json:"-" axe:""`
	Hello string
}

func TestGetMeta(t *testing.T) {
	meta := GetMeta(&testJob{})
	assert.Equal(t, &Meta{
		Type:   reflect.TypeOf(testJob{}),
		Name:   "test",
		Coding: stick.JSON,
		Accessor: &stick.Accessor{
			Name: "axe.testJob",
			Fields: map[string]*stick.Field{
				"Data": {
					Index: 1,
					Type: reflect.TypeOf(""),
				},
			},
		},
	}, meta)

	meta = GetMeta(&bsonJob{})
	assert.Equal(t, &Meta{
		Type:   reflect.TypeOf(bsonJob{}),
		Name:   "bson",
		Coding: stick.BSON,
		Accessor: &stick.Accessor{
			Name: "axe.bsonJob",
			Fields: map[string]*stick.Field{
				"Data": {
					Index: 1,
					Type: reflect.TypeOf(""),
				},
			},
		},
	}, meta)

	assert.PanicsWithValue(t, `axe: expected first struct field to be an embedded "axe.Base"`, func() {
		GetMeta(&invalidJob1{})
	})

	assert.PanicsWithValue(t, `axe: expected to find a coding tag of the form 'json:"-"' or 'bson:"-"' on "axe.Base"`, func() {
		GetMeta(&invalidJob2{})
	})

	assert.PanicsWithValue(t, `axe: expected to find a tag of the form 'axe:"name"' on "axe.Base"`, func() {
		GetMeta(&invalidJob3{})
	})
}

func TestMetaMake(t *testing.T) {
	job := GetMeta(&testJob{}).Make()
	assert.Equal(t, &testJob{}, job)
}

func TestDynamicAccess(t *testing.T) {
	job := &testJob{
		Data: "data",
	}

	val, ok := stick.Get(job, "data")
	assert.False(t, ok)
	assert.Nil(t, val)

	val, ok = stick.Get(job, "Data")
	assert.True(t, ok)
	assert.Equal(t, "data", val)

	ok = stick.Set(job, "data", "foo")
	assert.False(t, ok)
	assert.Equal(t, "data", job.Data)

	ok = stick.Set(job, "Data", "foo")
	assert.True(t, ok)
	assert.Equal(t, "foo", job.Data)
}
