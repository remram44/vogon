package database

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

type Locker interface {
	Lock()
	Unlock()
}

type KeyValueStore interface {
	Read(key string) (Object, error)
	Write(key string, value Object) error
	Delete(key string) error
}

type KvDatabase struct {
	mutex Locker
	store KeyValueStore
}

func NewKvDatabase(mutex Locker, store KeyValueStore) *KvDatabase {
	return &KvDatabase{
		mutex: mutex,
		store: store,
	}
}

func (db *KvDatabase) Create(object Object, replace bool) (MetadataResponse, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	previous, err := db.store.Read(object.Metadata.Name)
	exists := true
	if err != nil {
		if _, ok := err.(*DoesNotExist); ok {
			exists = false
		} else {
			return MetadataResponse{}, err
		}
	}
	if exists {
		if !replace {
			return MetadataResponse{}, &Conflict{
				s: fmt.Sprintf("Object %s already exists, cannot create", object.Metadata.Name),
			}
		}
		if object.Metadata.Id != "" {
			if previous.Metadata.Id != object.Metadata.Id {
				return MetadataResponse{}, &Conflict{
					s: fmt.Sprintf("Object %s exists and does not have the expected id, cannot replace", object.Metadata.Name),
				}
			}
		}

		if object.Metadata.Revision != "" {
			if object.Metadata.Id == "" {
				return MetadataResponse{}, errors.New("Cannot replace with a previous revision but no previous id")
			}
			if previous.Metadata.Revision != object.Metadata.Revision {
				return MetadataResponse{}, &Conflict{
					s: fmt.Sprintf("Object %s exists and does not have the expected revision, cannot replace", object.Metadata.Name),
				}
			}
		}

		object.Metadata.CreationTime = previous.Metadata.CreationTime
		object.Metadata.Id = previous.Metadata.Id
		object.Metadata.Revision = RandomString()
	} else {
		object.Metadata.CreationTime = time.Now()
		object.Metadata.Id = RandomString()
		object.Metadata.Revision = RandomString()
	}

	err = db.store.Write(object.Metadata.Name, object)
	if err != nil {
		return MetadataResponse{}, err
	}

	return MetadataResponse{
		Id:       object.Metadata.Id,
		Revision: object.Metadata.Revision,
	}, nil
}

func (db *KvDatabase) Update(object Object) (MetadataResponse, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	previous, err := db.store.Read(object.Metadata.Name)
	exists := true
	if err != nil {
		if _, ok := err.(*DoesNotExist); ok {
			exists = false
		} else {
			return MetadataResponse{}, err
		}
	}
	if !exists {
		return MetadataResponse{}, &DoesNotExist{
			s: fmt.Sprintf("Object %s does not exist, cannot update", object.Metadata.Name),
		}
	}

	if object.Metadata.Id != "" {
		if previous.Metadata.Id != object.Metadata.Id {
			return MetadataResponse{}, &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected id, cannot update", object.Metadata.Name),
			}
		}
	}

	if object.Metadata.Revision != "" {
		if object.Metadata.Id == "" {
			return MetadataResponse{}, errors.New("Cannot update with a previous revision but no previous id")
		}
		if previous.Metadata.Revision != object.Metadata.Revision {
			return MetadataResponse{}, &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected revision, cannot update", object.Metadata.Name),
			}
		}
	}

	object.Metadata.CreationTime = previous.Metadata.CreationTime
	object.Metadata.Id = previous.Metadata.Id
	object.Metadata.Revision = RandomString()

	err = db.store.Write(object.Metadata.Name, object)
	if err != nil {
		return MetadataResponse{}, err
	}

	return MetadataResponse{
		Id:       object.Metadata.Id,
		Revision: object.Metadata.Revision,
	}, nil
}

func (db *KvDatabase) Get(name string) (Object, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	object, err := db.store.Read(name)
	if err != nil {
		if _, ok := err.(*DoesNotExist); ok {
			return Object{}, &DoesNotExist{
				s: fmt.Sprintf("Object %s does not exist", name),
			}
		} else {
			return object, err
		}
	}

	return object, nil
}

func (db *KvDatabase) Delete(name string, id string, revision string) (MetadataResponse, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	previous, err := db.store.Read(name)
	if err != nil {
		if _, ok := err.(*DoesNotExist); ok {
			return MetadataResponse{}, &DoesNotExist{
				s: fmt.Sprintf("Object %s does not exist", name),
			}
		} else {
			return MetadataResponse{}, err
		}
	}

	if id != "" {
		if previous.Metadata.Id != id {
			return MetadataResponse{}, &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected id, cannot delete", name),
			}
		}
	}

	if revision != "" {
		if id == "" {
			return MetadataResponse{}, errors.New("Cannot delete with a previous revision but no previous id")
		}
		if previous.Metadata.Revision != revision {
			return MetadataResponse{}, &Conflict{
				s: fmt.Sprintf("Object %s does not have the expected revision, cannot delete", name),
			}
		}
	}

	err = db.store.Delete(name)
	if err != nil {
		return MetadataResponse{}, err
	}

	return MetadataResponse{
		Id:       previous.Metadata.Id,
		Revision: previous.Metadata.Revision,
	}, nil
}

func RandomString() string {
	num, err := rand.Int(rand.Reader, big.NewInt(0x100000000))
	if err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return fmt.Sprintf("%x", num)
}
