package fire

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/test"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Post struct {
	Base       `bson:",inline" fire:"posts"`
	Title      string  `json:"title" valid:"required" bson:"title" fire:"filterable,sortable"`
	Published  bool    `json:"published" valid:"-" fire:"filterable"`
	TextBody   string  `json:"text-body" valid:"-" bson:"text_body"`
	Comments   HasMany `json:"-" valid:"-" bson:"-" fire:"comments:comments:post"`
	Selections HasMany `json:"-" valid:"-" bson:"-" fire:"selections:selections:posts"`
}

type Comment struct {
	Base    `bson:",inline" fire:"comments"`
	Message string         `json:"message" valid:"required"`
	Parent  *bson.ObjectId `json:"-" valid:"-" fire:"parent:comments"`
	PostID  bson.ObjectId  `json:"-" valid:"required" bson:"post_id" fire:"post:posts"`
}

type Selection struct {
	Base    `bson:",inline" fire:"selections:selections"`
	Name    string          `json:"name" valid:"required"`
	PostIDs []bson.ObjectId `json:"-" valid:"-" bson:"post_ids" fire:"posts:posts"`
}

var session *mgo.Session

func init() {
	// connect to local mongodb
	sess, err := mgo.Dial("mongodb://0.0.0.0:27017/fire")
	if err != nil {
		panic(err)
	}

	// store session globally
	session = sess
}

func getDB() (*mgo.Session, *mgo.Database) {
	// get db
	db := session.DB("")

	// clean database by removing all documents
	db.C("posts").RemoveAll(nil)
	db.C("comments").RemoveAll(nil)
	db.C("selections").RemoveAll(nil)

	return session, db
}

func buildServer() (*echo.Echo, *mgo.Database) {
	// get db
	sess, db := getDB()

	// create router
	router := echo.New()

	// create set
	set := NewSet(sess, router, "")

	// add controllers
	set.Mount(&Controller{
		Model: &Post{},
	}, &Controller{
		Model: &Comment{},
	}, &Controller{
		Model: &Selection{},
	})

	// return router
	return router, db
}

func testRequest(e *echo.Echo, method, path string, headers map[string]string, payload string, callback func(*test.ResponseRecorder, engine.Request)) {
	req := test.NewRequest(method, path, strings.NewReader(payload))
	rec := test.NewResponseRecorder()

	for k, v := range headers {
		req.Header().Set(k, v)
	}

	e.ServeHTTP(req, rec)

	callback(rec, req)
}

func saveModel(db *mgo.Database, model Model) Model {
	Init(model)

	err := db.C(model.Meta().Collection).Insert(model)
	if err != nil {
		panic(err)
	}

	return model
}

func findLastModel(db *mgo.Database, model Model) Model {
	Init(model)

	err := db.C(model.Meta().Collection).Find(nil).Sort("-_id").One(model)
	if err != nil {
		panic(err)
	}

	return model
}
