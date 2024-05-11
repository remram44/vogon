package database

import (
	"time"
)

type Object struct {
	// Kind in URI format, e.g. "github.com/remram44/vogon/schemas/Job
	Kind string
	// Version identifier, can be any string e.g. "v1alpha1"
	Version  string
	Metadata ObjectMetadata
	// Spec: description of the desired state
	Spec any
	// Status: information about the current state
	Status any
}

type ObjectMetadata struct {
	Name   string
	Labels map[string]string
	//annotations???
	CreationTime time.Time
	// Unique string that is assigned on creation
	Id string
	// Unique string that changes every update
	Revision string
}

type MetadataResponse struct {
	Id       string
	Revision string
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
	// If replace is true and Id or Revision are not empty, returns an error if
	// they don't match.
	Create(object Object, replace bool) (MetadataResponse, error)

	// Update an existing object
	//
	// If Id or Revision are not empty, returns an error if they don't match.
	Update(object Object) (MetadataResponse, error)

	// Get a single object by name
	Get(name string) (Object, error)

	//List(???) ([]Object, error)

	// Delete an object
	//
	// If previousRevision is not empty, returns an error if it doesn't match
	// the revision on the server.
	Delete(name string, id string, revision string) (MetadataResponse, error)
}
