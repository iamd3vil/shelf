package main

import (
	"encoding/json"
	"io"
	"os"
	"path"
)

// DB is the db for shelf to store where each symlink is supposed to go.
// This is a JSON File generally stored in
type DB struct {
	Name  string            `json:"name"`
	Links map[string]string `json:"links"`
}

// Marshal marshals DB into JSON
func (db *DB) Marshal(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(&db)
	if err != nil {
		return err
	}
	return nil
}

// AddLink adds the file and link paths to the DB
func (db *DB) AddLink(filePath, linkPath string) {
	db.Links[filePath] = linkPath
}

// NewDB creates a shelf DB
func NewDB(shelfPath string, shelfName string) (*DB, error) {
	db := DB{
		Name:  shelfName,
		Links: make(map[string]string),
	}

	p := path.Join(shelfPath, "shelf.json")

	f, err := os.Create(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = db.Marshal(f)
	if err != nil {
		return nil, err
	}

	return &db, nil
}

// GetDB returns the DB in the given shelf
func GetDB(shelfPath string) (*DB, string, error) {
	dbPath := path.Join(shelfPath, "shelf.json")
	f, err := os.Open(dbPath)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	db := &DB{}
	err = json.NewDecoder(f).Decode(db)
	if err != nil {
		return nil, "", err
	}

	return db, dbPath, nil
}