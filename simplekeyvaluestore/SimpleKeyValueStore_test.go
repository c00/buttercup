package simplekeyvaluestore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleKeyValueStoreEdgeCases(t *testing.T) {
	// Test Empty string edge case
	ks := NewSimpleKeyValueStore[string]()
	err := ks.Add("", "foo")
	assert.Nil(t, err, "An empty string is also a string")

	val, err := ks.Get("")
	assert.Nil(t, err, "No error expected. An empty string is a valid key")
	assert.Equal(t, val, "foo", "Expected 'foo' as a value")
}

func TestSimpleKeyValueStoreStrings(t *testing.T) {
	ks := NewSimpleKeyValueStore[string]()

	assert.False(t, ks.Has("foo"), "Foo does not exist yet")

	_, err := ks.Get("foo")
	assert.Error(t, err, "Foo should not exist")

	err = ks.Add("foo", "bar")
	assert.Nil(t, err, "should be nil")

	assert.True(t, ks.Has("foo"), "Should have a foo now.")

	err = ks.Add("foo", "something else")
	assert.NotNil(t, err, "should not be nil, item already exists")

	gotten, err := ks.Get("foo")
	assert.Nil(t, err, "no error expected")
	assert.Equal(t, "bar", gotten, "foo should have the value bar")

	ks.Set("foo", "something else")

	gotten, err = ks.Get("foo")
	assert.Nil(t, err, "no error expected")
	assert.Equal(t, "something else", gotten, "foo should have the newly set value")

	assert.True(t, ks.Delete("foo"), "foo should be deleted")

	assert.False(t, ks.Has("foo"), "foo should be gone")
	assert.False(t, ks.Delete("foo"), "foo was already gone")
}

func TestSimpleKeyValueStoreStructs(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	ks := NewSimpleKeyValueStore[User]()

	_, err := ks.Get("foo")
	assert.Error(t, err, "Foo should not exist")

	err = ks.Add("foo", User{Name: "Peter", Age: 45})
	assert.Nil(t, err, "should be nil")

	gotten, err := ks.Get("foo")
	assert.Nil(t, err, "no error expected")
	assert.Equal(t, User{Name: "Peter", Age: 45}, gotten, "User should be the same")
}
