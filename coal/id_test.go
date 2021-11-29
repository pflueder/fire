package coal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestP(t *testing.T) {
	id := New()
	assert.Equal(t, &id, P(id))
}

func TestN(t *testing.T) {
	var id *ID
	assert.Equal(t, id, N())
	assert.NotEqual(t, nil, N())
}

func TestIsHex(t *testing.T) {
	assert.False(t, IsHex("foo"))
	assert.False(t, IsHex(""))
	assert.True(t, IsHex(New().Hex()))
}

func TestMustFromHex(t *testing.T) {
	assert.NotPanics(t, func() {
		MustFromHex(New().Hex())
	})

	assert.Panics(t, func() {
		MustFromHex("foo")
	})
}

func TestUnique(t *testing.T) {
	id1 := New()
	id2 := New()

	assert.Nil(t, Unique(nil))
	assert.Equal(t, []ID{id1}, Unique([]ID{id1}))
	assert.Equal(t, []ID{id1}, Unique([]ID{id1, id1}))
	assert.Equal(t, []ID{id1, id2}, Unique([]ID{id1, id2, id1}))
	assert.Equal(t, []ID{id1, id2}, Unique([]ID{id1, id2, id1, id2}))
}

func TestContains(t *testing.T) {
	a := New()
	b := New()
	c := New()
	d := New()

	assert.True(t, Contains([]ID{a, b, c}, a))
	assert.True(t, Contains([]ID{a, b, c}, b))
	assert.True(t, Contains([]ID{a, b, c}, c))
	assert.False(t, Contains([]ID{a, b, c}, d))
}

func TestIncludes(t *testing.T) {
	a := New()
	b := New()
	c := New()
	d := New()

	assert.True(t, Includes([]ID{a, b, c}, []ID{a}))
	assert.True(t, Includes([]ID{a, b, c}, []ID{a, b}))
	assert.True(t, Includes([]ID{a, b, c}, []ID{a, b, c}))
	assert.False(t, Includes([]ID{a, b, c}, []ID{a, b, c, d}))
	assert.False(t, Includes([]ID{a, b, c}, []ID{d}))
}

func TestUnion(t *testing.T) {
	a := New()
	b := New()
	c := New()
	d := New()
	e := New()

	assert.Nil(t, Union())
	assert.Nil(t, Union(nil))
	assert.Equal(t, []ID{a}, Union(nil, []ID{a}))
	assert.Equal(t, []ID{b, a, c, d, e}, Union([]ID{b, a, c}, []ID{d, a, b}, []ID{e}))
}

func TestSubtract(t *testing.T) {
	a := New()
	b := New()
	c := New()

	assert.Nil(t, Subtract(nil, nil))
	assert.Nil(t, Subtract(nil, []ID{a, b}))
	assert.Equal(t, []ID{a, b}, Subtract([]ID{a, b}, nil))
	assert.Equal(t, []ID{a}, Subtract([]ID{a, b}, []ID{b, c}))
}

func TestIntersect(t *testing.T) {
	a := New()
	b := New()
	c := New()

	assert.Nil(t, Intersect(nil, nil))
	assert.Nil(t, Intersect(nil, []ID{a, b}))
	assert.Nil(t, Intersect([]ID{a, b}, nil))
	assert.Equal(t, []ID{b}, Intersect([]ID{a, b}, []ID{b, c}))
}
