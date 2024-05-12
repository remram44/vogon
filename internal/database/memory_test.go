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

		meta, err := db.Create(
			Object{
				Kind:    "example.org/Example",
				Version: "v1",
				Metadata: ObjectMetadata{
					Name:     "one",
					Id:       "12345",
					Revision: "67890",
				},
				Spec:   FakeSpec{value: "yay"},
				Status: struct{}{},
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
		if object.Metadata.Id != meta.Id ||
			object.Metadata.Revision != meta.Revision {
			t.Fatalf("MetadataResponse invalid")
		}
		if object.Metadata.Name != "one" ||
			object.Metadata.Id == "12345" ||
			object.Metadata.Id == "" ||
			object.Metadata.Revision == "67890" ||
			object.Metadata.Revision == "" {
			t.Fatalf("object has invalid metadata: %#v", object.Metadata)
		}
		if object.Spec.(FakeSpec).value != "yay" {
			t.Fatal("object has invalid spec")
		}
	}
}

func TestCreateReplace(t *testing.T) {
	db := NewInMemoryDatabase()
	var _ Database = db

	// Put in a first object, will be replaced
	meta, err := db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       "12345",
				Revision: "67890",
			},
			Spec:   struct{}{},
			Status: struct{}{},
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

	// Replace with wrong ID
	meta, err = db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       "12345",
				Revision: "",
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
		true,
	)
	if err == nil {
		t.Fatal("replace with wrong id didn't fail")
	}
	if _, ok := err.(*Conflict); !ok {
		t.Fatal("replace with wrong id didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("replace failed but MetadataResponse is set")
	}

	// Replace with wrong revision
	meta, err = db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       previous.Metadata.Id,
				Revision: "567890",
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
		true,
	)
	if err == nil {
		t.Fatal("replace with wrong revision didn't fail")
	}
	if _, ok := err.(*Conflict); !ok {
		t.Fatal("replace with wrong revision didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("replace failed but MetadataResponse is set")
	}

	// Replace with revision but no id
	meta, err = db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       "",
				Revision: previous.Metadata.Revision,
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
		true,
	)
	if err == nil || err.Error() != "Cannot replace with a previous revision but no previous id" {
		t.Fatal("replace with revision but no id didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("replace failed but MetadataResponse is set")
	}

	// Replace with id and revision
	meta, err = db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       previous.Metadata.Id,
				Revision: previous.Metadata.Revision,
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
		true,
	)
	if err != nil {
		t.Fatal("replace with correct id and revision didn't work")
	}
	if db.objects["one"].Metadata.Id != meta.Id ||
		db.objects["one"].Metadata.Revision != meta.Revision {
		t.Fatalf("MetadataResponse invalid")
	}

	// Replace with id
	meta, err = db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name: "one",
				Id:   previous.Metadata.Id,
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
		true,
	)
	if err != nil {
		t.Fatal("replace with correct id didn't work")
	}
	if db.objects["one"].Metadata.Id != meta.Id ||
		db.objects["one"].Metadata.Revision != meta.Revision {
		t.Fatalf("MetadataResponse invalid")
	}

	// Replace with no previous fields
	meta, err = db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name: "one",
			},
			Spec:   struct{}{},
			Status: struct{}{},
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
	if object.Metadata.Id != meta.Id ||
		object.Metadata.Revision != meta.Revision {
		t.Fatalf("MetadataResponse invalid")
	}
	if object.Metadata.Name != "one" ||
		object.Metadata.Id == "12345" ||
		object.Metadata.Id == "" ||
		object.Metadata.Revision == "67890" ||
		object.Metadata.Revision == "" {
		t.Fatalf("object has invalid metadata: %#v", object.Metadata)
	}
	if object.Metadata.Id != previous.Metadata.Id {
		t.Fatal("object id changed")
	}
	if object.Metadata.Revision == previous.Metadata.Revision {
		t.Fatal("object revision didn't change")
	}
}

func TestUpdate(t *testing.T) {
	db := NewInMemoryDatabase()
	var _ Database = db

	// Update missing object
	meta, err := db.Update(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name: "one",
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
	)
	if err == nil {
		t.Fatal("update missing object didn't fail")
	}
	if _, ok := err.(*DoesNotExist); !ok {
		t.Fatal("update missing object didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("update failed but MetadataResponse is set")
	}

	// Put in a first object, will be updated
	meta, err = db.Create(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       "12345",
				Revision: "67890",
			},
			Spec:   struct{}{},
			Status: struct{}{},
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
	meta, err = db.Update(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       "12345",
				Revision: previous.Metadata.Revision,
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
	)
	if err == nil {
		t.Fatal("update with wrong id didn't fail")
	}
	if _, ok := err.(*Conflict); !ok {
		t.Fatal("update with wrong id didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("update failed but MetadataResponse is set")
	}

	// Update with wrong revision
	meta, err = db.Update(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       previous.Metadata.Id,
				Revision: "567890",
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
	)
	if err == nil {
		t.Fatal("update with wrong revision didn't fail")
	}
	if _, ok := err.(*Conflict); !ok {
		t.Fatal("update with wrong revision didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("update failed but MetadataResponse is set")
	}

	// Update with revision but no id
	meta, err = db.Update(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       "",
				Revision: previous.Metadata.Revision,
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
	)
	if err == nil || err.Error() != "Cannot update with a previous revision but no previous id" {
		t.Fatal("update with revision but no id didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("update failed but MetadataResponse is set")
	}

	// Update with id and revision
	meta, err = db.Update(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name:     "one",
				Id:       previous.Metadata.Id,
				Revision: previous.Metadata.Revision,
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
	)
	if err != nil {
		t.Fatal("update with correct id and revision didn't work")
	}
	if db.objects["one"].Metadata.Id != meta.Id ||
		db.objects["one"].Metadata.Revision != meta.Revision {
		t.Fatalf("MetadataResponse invalid")
	}

	// Update with id
	meta, err = db.Update(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name: "one",
				Id:   previous.Metadata.Id,
			},
			Spec:   struct{}{},
			Status: struct{}{},
		},
	)
	if err != nil {
		t.Fatal("update with correct id didn't work")
	}
	if db.objects["one"].Metadata.Id != meta.Id ||
		db.objects["one"].Metadata.Revision != meta.Revision {
		t.Fatalf("MetadataResponse invalid")
	}

	// Update with no previous fields
	meta, err = db.Update(
		Object{
			Kind:    "example.org/Example",
			Version: "v1",
			Metadata: ObjectMetadata{
				Name: "one",
			},
			Spec:   FakeSpec{value: "yay"},
			Status: struct{}{},
		},
	)
	if err != nil {
		t.Fatal("update with no comparison didn't work")
	}
	if db.objects["one"].Metadata.Id != meta.Id ||
		db.objects["one"].Metadata.Revision != meta.Revision {
		t.Fatalf("MetadataResponse invalid")
	}

	object, err := db.Get("one")
	if err != nil {
		t.Fatal("get didn't work")
	}
	if object.Metadata.Name != "one" ||
		object.Metadata.Id == "" ||
		object.Metadata.Revision == "" {
		t.Fatalf("object has invalid metadata: %#v", object.Metadata)
	}
	if object.Spec.(FakeSpec).value != "yay" {
		t.Fatal("object has invalid spec")
	}
}

func TestDelete(t *testing.T) {
	getDb := func() *InMemoryDatabase {
		db := NewInMemoryDatabase()
		var _ Database = db

		// Put in a first object, will be deleted
		_, err := db.Create(
			Object{
				Kind:    "example.org/Example",
				Version: "v1",
				Metadata: ObjectMetadata{
					Name:     "one",
					Id:       "12345",
					Revision: "67890",
				},
				Spec:   struct{}{},
				Status: struct{}{},
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

	meta, err := db.Delete("one", "123456", "")
	if err == nil {
		t.Fatal("delete didn't fail")
	}
	if _, ok := err.(*DoesNotExist); !ok {
		t.Fatal("delete didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("delete failed but MetadataResponse is set")
	}

	meta, err = db.Delete("one", "", "")
	if err == nil {
		t.Fatal("delete didn't fail")
	}
	if _, ok := err.(*DoesNotExist); !ok {
		t.Fatal("delete didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("delete failed but MetadataResponse is set")
	}

	// Delete by name
	db = getDb()
	meta, err = db.Delete("one", "", "")
	if err != nil {
		t.Fatal("delete by name didn't work")
	}

	// Delete with wrong id
	db = getDb()
	meta, err = db.Delete("one", "12345", "")
	if err == nil {
		t.Fatal("delete with wrong id didn't fail")
	}
	if _, ok := err.(*Conflict); !ok {
		t.Fatal("delete with wrong id didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("delete failed but MetadataResponse is set")
	}

	// Delete with id
	db = getDb()
	meta, err = db.Delete("one", db.objects["one"].Metadata.Id, "")
	if err != nil {
		t.Fatal("delete with id didn't work")
	}

	// Delete with revision but no id
	db = getDb()
	meta, err = db.Delete("one", "", db.objects["one"].Metadata.Revision)
	if err == nil || err.Error() != "Cannot delete with a previous revision but no previous id" {
		t.Fatal("delete with revision but no id didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("delete failed but MetadataResponse is set")
	}

	// Delete with wrong revision
	db = getDb()
	meta, err = db.Delete("one", db.objects["one"].Metadata.Id, "4567")
	if err == nil {
		t.Fatal("delete with wrong revision didn't fail")
	}
	if _, ok := err.(*Conflict); !ok {
		t.Fatal("delete with wrong revision didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("delete failed but MetadataResponse is set")
	}

	// Delete with revision
	db = getDb()
	meta, err = db.Delete("one", db.objects["one"].Metadata.Id, db.objects["one"].Metadata.Revision)
	if err != nil {
		t.Fatal("delete with revision didn't work")
	}
}
