package database

import (
	"time"
)

type Object struct {
	// Kind in URI format, e.g. "github.com/remram44/vogon/schemas/Job
	kind string
	// Version identifier, can be any string e.g. "v1alpha1"
	version  string
	metadata ObjectMetadata
	// Spec: description of the desired state
	spec any
	// Status: information about the current state
	status any
}

type ObjectMetadata struct {
	name   string
	labels map[string]string
	//annotations???
	creationTime time.Time
	// Unique string that is assigned on creation
	id string
	// Unique string that changes every update
	revision string
}

type Conflict struct {
	s string
}

func (e *Conflict) Error() string {
	return e.s
}

type DoesNotExist struct {
	s string
}

func (e *DoesNotExist) Error() string {
	return e.s
}

type Database interface {
	// Create an object
	//
	// If replace is false, returns an error if it exists.
	Create(object Object, replace bool) error

	// Update an existing object
	//
	// If metadata.revision is not empty, returns an error if it doesn't
	// match the revision on the server.
	Update(object Object) error

	// Get a single object by name
	Get(name string) (Object, error)

	//List(???) ([]Object, error)

	// Delete an object
	//
	// If previousRevision is not empty, returns an error if it doesn't match
	// the revision on the server.
	Delete(name string, id string, revision string) error
}
