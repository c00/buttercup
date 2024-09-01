package simplekeyvaluestore

import (
	"errors"
	"sync"
)

// NewSimpleKeyValueStore Creates a new instance of SimpleKeyValueStore
func NewSimpleKeyValueStore[T any]() SimpleKeyValueStore[T] {
	return SimpleKeyValueStore[T]{
		data: make(map[string]T),
	}
}

// SimpleKeyValueStore is a threadsafe key value store.
type SimpleKeyValueStore[T any] struct {
	data map[string]T
	mu   sync.RWMutex
}

// Add set the value. If it already exists, returns an error
// This store does not care what the key looks like, as long as it's a string.
func (ks *SimpleKeyValueStore[T]) Add(key string, value T) error {
	if ks.Has(key) {
		return errors.New("key is not unique")
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.data[key] = value

	return nil
}

// Set the value. If it already exists, it gets overridden.
// This store does not care what the key looks like, as long as it's a string.
func (ks *SimpleKeyValueStore[T]) Set(key string, value T) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.data[key] = value
}

// Delete deletes a key
func (ks *SimpleKeyValueStore[T]) Delete(key string) bool {
	if !ks.Has(key) {
		return false
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()

	delete(ks.data, key)
	return true
}

// Has checks if the store has a key
func (ks *SimpleKeyValueStore[T]) Has(key string) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	_, ok := ks.data[key]

	return ok
}

// Get value from store. If it doesn't exist, throw an error
func (ks *SimpleKeyValueStore[T]) Get(key string) (T, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	value, exists := ks.data[key]
	if exists {
		return value, nil
	}
	return value, errors.New("key does not exist")
}
