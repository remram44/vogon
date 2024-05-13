package database

import (
	"sync"
)

type mutexLocker struct {
	mutex sync.Mutex
}

func (l *mutexLocker) Lock() {
	l.mutex.Lock()
}

func (l *mutexLocker) Unlock() {
	l.mutex.Unlock()
}

type inMemoryKv struct {
	objects map[string]Object
}

func (m *inMemoryKv) Read(key string) (Object, error) {
	object, exists := m.objects[key]
	if !exists {
		return object, &DoesNotExist{
			s: "no such key in memory",
		}
	}
	return object, nil
}

func (m *inMemoryKv) Write(key string, value Object) error {
	m.objects[key] = value
	return nil
}

func (m *inMemoryKv) Delete(key string) error {
	_, exists := m.objects[key]
	if !exists {
		return &DoesNotExist{
			s: "no such key in memory",
		}
	}

	delete(m.objects, key)
	return nil
}

func NewInMemoryDatabase() *KvDatabase {
	return NewKvDatabase(
		&mutexLocker{},
		&inMemoryKv{
			objects: make(map[string]Object),
		},
	)
}
