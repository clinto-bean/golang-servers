package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

var ErrNotExist = errors.New("resource does not exist")

type DB struct {
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]User `json:"users"`
	Tokens map[string]Token `json:"tokens"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Email string
	Password string
	ID int
}

type Token struct {
	ID int 
	Body string `json:"token"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users: map[int]User{},
		Tokens: map[string]Token{},
	}

	db.mu.RLock()
	err := db.writeDB(dbStructure)
	db.mu.RUnlock()

	return err
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	dbStructure := DBStructure{}
	
	db.mu.RLock()
	dat, err := os.ReadFile(db.path)
	db.mu.RUnlock()

	if errors.Is(err, os.ErrNotExist) {
		log.Printf("Could not read file: %v", db.path)
		return dbStructure, err
	}

	err = json.Unmarshal(dat, &dbStructure)

	if err != nil {
		log.Printf("Could not unmarshal data: %v", dat)
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		log.Printf("Could not marshal data: %v", dat)
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		log.Printf("Could not write new data to db: %v", err)
		return err
	}
	return nil
}



