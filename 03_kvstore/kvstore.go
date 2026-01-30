package main

import (
	"math/rand/v2"
	"sync"
	"time"
)

type KVStore struct {
	mu    sync.Mutex
	Store map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{Store: make(map[string]string)}
}

func (kvs *KVStore) Set(key string, value string) {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()

	time.Sleep(time.Duration(rand.IntN(5)) * time.Millisecond)
	kvs.Store[key] = value
}

func (kvs *KVStore) Get(key string) (string, bool) {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()

	value, exists := kvs.Store[key]
	time.Sleep(time.Duration(rand.IntN(5)) * time.Millisecond)

	return value, exists
}
