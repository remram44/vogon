package database

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type InMemoryDatabase struct {
	mutex   sync.Mutex
	objects map[string]Object
}

func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		objects: make(map[string]Object),
	}
}

func (db *InMemoryDatabase) Create(object Object, replace bool) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	previous, exists := db.objects[object.metadata.name]
	if exists {
		if !replace {
			return &Conflict{
				s: fmt.Sprintf("Object %s already exists, cannot create", object.metadata.name),
			}
		}
		object.metadata.creationTime = previous.metadata.creationTime
		object.metadata.id = previous.metadata.id
		object.metadata.revision = RandomString()
	} else {
		object.metadata.creationTime = time.Now()
		object.metadata.id = RandomString()
		object.metadata.revision = RandomString()
	}

	db.objects[object.metadata.name] = object

	return nil
}

func (db *InMemoryDatabase) Update(object Object) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	previous, exists := db.objects[object.metadata.name]
	if !exists {
		return &DoesNotExist{
			s: fmt.Sprintf("Object %s does not exist, cannot update", object.metadata.name),
		}
	}

	if object.metadata.id != "" {
		if previous.metadata.id != object.metadata.id {
			return &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected id, cannot update", object.metadata.name),
			}
		}
	}

	if object.metadata.revision != "" {
		if object.metadata.id == "" {
			return errors.New("Cannot update with a previous revision but no previous id")
		}
		if previous.metadata.revision != object.metadata.revision {
			return &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected revision, cannot update", object.metadata.name),
			}
		}
	}

	object.metadata.creationTime = previous.metadata.creationTime
	object.metadata.id = previous.metadata.id
	object.metadata.revision = RandomString()

	db.objects[object.metadata.name] = object

	return nil
}

func (db *InMemoryDatabase) Get(name string) (Object, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	object, exists := db.objects[name]
	if !exists {
		return Object{}, &DoesNotExist{
			s: fmt.Sprintf("Object %s does not exist", name),
		}
	}

	return object, nil
}

func (db *InMemoryDatabase) Delete(name string, id string, revision string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	previous, exists := db.objects[name]
	if !exists {
		return &DoesNotExist{
			s: fmt.Sprintf("Object %s does not exist", name),
		}
	}

	if id != "" {
		if previous.metadata.id != id {
			return &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected id, cannot delete", name),
			}
		}
	}

	if revision != "" {
		if id == "" {
			return errors.New("Cannot delete with a previous revision but no previous id")
		}
		if previous.metadata.revision != revision {
			return &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected revision, cannot delete", name),
			}
		}
	}

	delete(db.objects, name)

	return nil
}

func RandomString() string {
	num, err := rand.Int(rand.Reader, big.NewInt(0x100000000))
	if err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return fmt.Sprintf("%x", num)
}
