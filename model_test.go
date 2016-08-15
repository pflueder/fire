package fire

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDHelper(t *testing.T) {
	post := Init(&Post{}).(*Post)
	assert.Equal(t, post.DocID, post.ID())
}

func TestGetHelper(t *testing.T) {
	post1 := Init(&Post{})
	assert.Equal(t, "", post1.Get("text_body"))
	assert.Equal(t, "", post1.Get("text-body"))
	assert.Equal(t, "", post1.Get("TextBody"))

	post2 := Init(&Post{TextBody: "hello"})
	assert.Equal(t, "hello", post2.Get("text_body"))
	assert.Equal(t, "hello", post2.Get("text-body"))
	assert.Equal(t, "hello", post2.Get("TextBody"))

	assert.Panics(t, func() {
		post1.Get("missing")
	})
}

func TestSetHelper(t *testing.T) {
	post := Init(&Post{}).(*Post)

	post.Set("text_body", "1")
	assert.Equal(t, "1", post.TextBody)

	post.Set("text-body", "2")
	assert.Equal(t, "2", post.TextBody)

	post.Set("TextBody", "3")
	assert.Equal(t, "3", post.TextBody)

	assert.Panics(t, func() {
		post.Set("missing", "-")
	})

	assert.Panics(t, func() {
		post.Set("TextBody", 1)
	})
}
