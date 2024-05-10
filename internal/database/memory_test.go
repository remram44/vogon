package database

import (
	"testing"
)

type FakeSpec struct {
	value string
}

func TestCreate(t *testing.T) {
	for _, replace := range []bool{false, true} {
		db := NewInMemoryDatabase()
		var _ Database = db

		err := db.Create(
			Object{
				kind:    "example.org/Example",
				version: "v1",
				metadata: ObjectMetadata{
					name:     "one",
					id:       "12345",
					revision: "67890",
				},
				spec:   FakeSpec{value: "yay"},
				status: struct{}{},
			},
			replace,
		)
		if err != nil {
			t.Fatalf("%#v", err)
		}

		if len(db.objects) != 1 {
			t.Fatal("invalid objects in database")
		}
		object, exists := db.objects["one"]
		if !exists {
			t.Fatal("object not in database")
		}
		if object.metadata.name != "one" ||
			object.metadata.id == "12345" ||
			object.metadata.id == "" ||
			object.metadata.revision == "67890" ||
			object.metadata.revision == "" {
			t.Fatalf("object has invalid metadata: %#v", object.metadata)
		}
		if object.spec.(FakeSpec).value != "yay" {
			t.Fatal("object has invalid spec")
		}
	}
}

func TestCreateReplace(t *testing.T) {
	db := NewInMemoryDatabase()
	var _ Database = db

	// Put in a first object, will be replaced
	err := db.Create(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name:     "one",
				id:       "12345",
				revision: "67890",
			},
			spec:   struct{}{},
			status: struct{}{},
		},
		false,
	)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	previous, exists := db.objects["one"]
	if !exists {
		t.Fatal("object not in database")
	}

	// Replace
	err = db.Create(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name:     "one",
				id:       "12345",
				revision: "67890",
			},
			spec:   struct{}{},
			status: struct{}{},
		},
		true,
	)
	if err != nil {
		t.Fatalf("%#v", err)
	}

	if len(db.objects) != 1 {
		t.Fatal("invalid objects in database")
	}
	object, exists := db.objects["one"]
	if !exists {
		t.Fatal("object not in database")
	}
	if object.metadata.name != "one" ||
		object.metadata.id == "12345" ||
		object.metadata.id == "" ||
		object.metadata.revision == "67890" ||
		object.metadata.revision == "" {
		t.Fatalf("object has invalid metadata: %#v", object.metadata)
	}
	if object.metadata.id != previous.metadata.id {
		t.Fatal("object id changed")
	}
	if object.metadata.revision == previous.metadata.revision {
		t.Fatal("object revision didn't change")
	}
}

func TestUpdate(t *testing.T) {
	db := NewInMemoryDatabase()
	var _ Database = db

	// Update missing object
	err := db.Update(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name: "one",
			},
			spec:   struct{}{},
			status: struct{}{},
		},
	)
	if err == nil || err.(*DoesNotExist) == nil {
		t.Fatal("update didn't raise DoesNotExist")
	}

	// Put in a first object, will be updated
	err = db.Create(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name:     "one",
				id:       "12345",
				revision: "67890",
			},
			spec:   struct{}{},
			status: struct{}{},
		},
		false,
	)
	if err != nil {
		t.Fatalf("%#v", err)
	}
	previous, exists := db.objects["one"]
	if !exists {
		t.Fatal("object not in database")
	}

	// Update with wrong ID
	err = db.Update(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name:     "one",
				id:       "12345",
				revision: previous.metadata.revision,
			},
			spec:   struct{}{},
			status: struct{}{},
		},
	)
	if err == nil || err.(*Conflict) == nil {
		t.Fatal("update with wrong id didn't fail")
	}

	// Update with wrong revision
	err = db.Update(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name:     "one",
				id:       previous.metadata.id,
				revision: "567890",
			},
			spec:   struct{}{},
			status: struct{}{},
		},
	)
	if err == nil || err.(*Conflict) == nil {
		t.Fatal("update with wrong revision didn't fail")
	}

	// Update with revision but no id
	err = db.Update(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name:     "one",
				id:       "",
				revision: previous.metadata.revision,
			},
			spec:   struct{}{},
			status: struct{}{},
		},
	)
	if err == nil || err.Error() != "Cannot update with a previous revision but no previous id" {
		t.Fatal("update with revision but no id didn't fail")
	}

	// Update with id and revision
	err = db.Update(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name:     "one",
				id:       previous.metadata.id,
				revision: previous.metadata.revision,
			},
			spec:   struct{}{},
			status: struct{}{},
		},
	)
	if err != nil {
		t.Fatal("update with correct id and revision didn't work")
	}

	// Update with id
	err = db.Update(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name: "one",
				id:   previous.metadata.id,
			},
			spec:   struct{}{},
			status: struct{}{},
		},
	)
	if err != nil {
		t.Fatal("update with correct id didn't work")
	}

	// Update with no previous fields
	err = db.Update(
		Object{
			kind:    "example.org/Example",
			version: "v1",
			metadata: ObjectMetadata{
				name: "one",
			},
			spec:   FakeSpec{value: "yay"},
			status: struct{}{},
		},
	)
	if err != nil {
		t.Fatal("update with no comparison didn't work")
	}

	object, err := db.Get("one")
	if err != nil {
		t.Fatal("get didn't work")
	}
	if object.metadata.name != "one" ||
		object.metadata.id == "" ||
		object.metadata.revision == "" {
		t.Fatalf("object has invalid metadata: %#v", object.metadata)
	}
	if object.spec.(FakeSpec).value != "yay" {
		t.Fatal("object has invalid spec")
	}
}

func TestDelete(t *testing.T) {
	getDb := func() *InMemoryDatabase {
		db := NewInMemoryDatabase()
		var _ Database = db

		// Put in a first object, will be deleted
		err := db.Create(
			Object{
				kind:    "example.org/Example",
				version: "v1",
				metadata: ObjectMetadata{
					name:     "one",
					id:       "12345",
					revision: "67890",
				},
				spec:   struct{}{},
				status: struct{}{},
			},
			false,
		)
		if err != nil {
			t.Fatalf("%#v", err)
		}
		_, exists := db.objects["one"]
		if !exists {
			t.Fatal("object not in database")
		}

		return db
	}

	// Delete missing object
	db := NewInMemoryDatabase()
	var _ Database = db

	err := db.Delete("one", "123456", "")
	if err == nil || err.(*DoesNotExist) == nil {
		t.Fatal("delete didn't raise DoesNotExist")
	}

	err = db.Delete("one", "", "")
	if err == nil || err.(*DoesNotExist) == nil {
		t.Fatal("delete didn't raise DoesNotExist")
	}

	// Delete by name
	db = getDb()
	err = db.Delete("one", "", "")
	if err != nil {
		t.Fatal("delete by name didn't work")
	}

	// Delete with wrong id
	db = getDb()
	err = db.Delete("one", "12345", "")
	if err == nil || err.(*Conflict) == nil {
		t.Fatal("delete with wrong id didn't fail")
	}

	// Delete with id
	db = getDb()
	err = db.Delete("one", db.objects["one"].metadata.id, "")
	if err != nil {
		t.Fatal("delete with id didn't work")
	}

	// Delete with revision but no id
	db = getDb()
	err = db.Delete("one", "", db.objects["one"].metadata.revision)
	if err == nil || err.Error() != "Cannot delete with a previous revision but no previous id" {
		t.Fatal("delete with revision but no id didn't fail")
	}

	// Delete with wrong revision
	db = getDb()
	err = db.Delete("one", db.objects["one"].metadata.id, "4567")
	if err == nil || err.(*Conflict) == nil {
		t.Fatal("delete with wrong id didn't fail")
	}

	// Delete with revision
	db = getDb()
	err = db.Delete("one", db.objects["one"].metadata.id, db.objects["one"].metadata.revision)
	if err != nil {
		t.Fatal("delete with id didn't work")
	}
}
