package database

import (
	"testing"
)

func fakeSpec(value string) interface{} {
	result := make(map[string]interface{})
	result["value"] = value
	return result
}

func runWithAllDatabases(t *testing.T, testFunc func(func(*testing.T) *KvDatabase, *testing.T)) {
	emptyInMemoryDb := func(t *testing.T) *KvDatabase {
		return NewInMemoryDatabase()
	}

	emptyFilesDb := func(t *testing.T) *KvDatabase {
		db, err := NewFilesDatabase(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		return db
	}

	t.Run("inmemory", func(t *testing.T) { testCreate(emptyInMemoryDb, t) })
	t.Run("files", func(t *testing.T) { testCreate(emptyFilesDb, t) })
}

func TestCreate(t *testing.T) {
	runWithAllDatabases(t, testCreate)
}

func testCreate(emptyDb func(t *testing.T) *KvDatabase, t *testing.T) {
	for _, replace := range []bool{false, true} {
		db := emptyDb(t)
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
				Spec:   fakeSpec("yay"),
				Status: struct{}{},
			},
			replace,
		)
		if err != nil {
			t.Fatalf("%#v", err)
		}

		object, err := db.Get("one")
		if err != nil {
			t.Fatalf("%#v", err)
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
		if object.Spec.(map[string]interface{})["value"] != "yay" {
			t.Fatal("object has invalid spec")
		}
	}
}

func TestCreateReplace(t *testing.T) {
	runWithAllDatabases(t, testCreateReplace)
}

func testCreateReplace(emptyDb func(t *testing.T) *KvDatabase, t *testing.T) {
	db := emptyDb(t)
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
	previous, err := db.Get("one")
	if err != nil {
		t.Fatalf("%#v", err)
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
	object, err := db.Get("one")
	if err != nil {
		t.Fatal("replace with correct id and revision didn't work")
	}
	if object.Metadata.Id != meta.Id ||
		object.Metadata.Revision != meta.Revision {
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
	object, err = db.Get("one")
	if err != nil {
		t.Fatal("replace with correct id didn't work")
	}
	if object.Metadata.Id != meta.Id ||
		object.Metadata.Revision != meta.Revision {
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

	object, err = db.Get("one")
	if err != nil {
		t.Fatal("replace didn't work")
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
	runWithAllDatabases(t, testUpdate)
}

func testUpdate(emptyDb func(t *testing.T) *KvDatabase, t *testing.T) {
	db := emptyDb(t)
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
	previous, err := db.Get("one")
	if err != nil {
		t.Fatal("create didn't work")
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
	object, err := db.Get("one")
	if err != nil {
		t.Fatal("update with correct id and revision didn't work")
	}
	if object.Metadata.Id != meta.Id ||
		object.Metadata.Revision != meta.Revision {
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
	object, err = db.Get("one")
	if err != nil {
		t.Fatal("update with correct id didn't work")
	}
	if object.Metadata.Id != meta.Id ||
		object.Metadata.Revision != meta.Revision {
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
			Spec:   fakeSpec("yay"),
			Status: struct{}{},
		},
	)
	if err != nil {
		t.Fatal("update with no comparison didn't work")
	}
	object, err = db.Get("one")
	if err != nil {
		t.Fatal("update with no comparison didn't work")
	}
	if object.Metadata.Id != meta.Id ||
		object.Metadata.Revision != meta.Revision {
		t.Fatalf("MetadataResponse invalid")
	}

	object, err = db.Get("one")
	if err != nil {
		t.Fatal("get didn't work")
	}
	if object.Metadata.Name != "one" ||
		object.Metadata.Id == "" ||
		object.Metadata.Revision == "" {
		t.Fatalf("object has invalid metadata: %#v", object.Metadata)
	}
	if object.Spec.(map[string]interface{})["value"] != "yay" {
		t.Fatal("object has invalid spec")
	}
}

func TestDelete(t *testing.T) {
	runWithAllDatabases(t, testDelete)
}

func testDelete(emptyDb func(t *testing.T) *KvDatabase, t *testing.T) {
	getDb := func() (*KvDatabase, MetadataResponse) {
		db := emptyDb(t)
		var _ Database = db

		// Put in a first object, will be deleted
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
		_, err = db.Get("one")
		if err != nil {
			t.Fatal("create didn't work")
		}

		return db, meta
	}

	// Delete missing object
	db := emptyDb(t)
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
	db, _ = getDb()
	meta, err = db.Delete("one", "", "")
	if err != nil {
		t.Fatal("delete by name didn't work")
	}

	// Delete with wrong id
	db, _ = getDb()
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
	db, createMeta := getDb()
	meta, err = db.Delete("one", createMeta.Id, "")
	if err != nil {
		t.Fatal("delete with id didn't work")
	}

	// Delete with revision but no id
	db, createMeta = getDb()
	meta, err = db.Delete("one", "", createMeta.Revision)
	if err == nil || err.Error() != "Cannot delete with a previous revision but no previous id" {
		t.Fatal("delete with revision but no id didn't fail")
	}
	if meta.Id != "" || meta.Revision != "" {
		t.Fatal("delete failed but MetadataResponse is set")
	}

	// Delete with wrong revision
	db, createMeta = getDb()
	meta, err = db.Delete("one", createMeta.Id, "4567")
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
	db, createMeta = getDb()
	meta, err = db.Delete("one", createMeta.Id, createMeta.Revision)
	if err != nil {
		t.Fatal("delete with revision didn't work")
	}
}
