package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"syscall"
)

type fileLocker struct {
	file *os.File
}

func (l *fileLocker) Lock() {
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX)
}

func (l *fileLocker) Unlock() {
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
}

type directoryKv struct {
	directory string
}

func (db *directoryKv) Read(name string) (Object, error) {
	var object Object

	file, err := os.Open(path.Join(db.directory, name+".json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return object, &DoesNotExist{
				s: "No such file",
			}
		}
		return object, err
	}
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&object)
	if err != nil {
		return object, err
	}
	return object, nil
}

func (db *directoryKv) Write(name string, object Object) error {
	filePath := path.Join(db.directory, name+".json")
	parentPath := path.Dir(filePath)
	os.MkdirAll(parentPath, 0666)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	return encoder.Encode(object)
}

func (db *directoryKv) Delete(name string) error {
	err := os.Remove(path.Join(db.directory, name+".json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &DoesNotExist{
				s: "No such file",
			}
		}
		return err
	}
	return nil
}

func NewFilesDatabase(directory string) (*KvDatabase, error) {
	lockFile, err := os.Create(path.Join(directory, "_lock"))
	if err != nil {
		return nil, fmt.Errorf("Error opening lock file: %w", err)
	}
	return NewKvDatabase(
		&fileLocker{
			file: lockFile,
		},
		&directoryKv{
			directory: directory,
		},
	), nil
}
