package database

import (
	"errors"
	"log"
)

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		log.Print("Could not load db.")
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:     id,
		Body:   body,
		Author: authorID,
	}
	dbStructure.Chirps[id] = chirp

	db.mu.Lock()
	err = db.writeDB(dbStructure)
	db.mu.Unlock()

	if err != nil {
		log.Print("Could not write db.")
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		log.Print("Could not get chirps.")
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	chirp, ok := dbStructure.Chirps[id]
	if !ok {
		return Chirp{}, ErrNotExist
	}
	return chirp, nil
}

func (db *DB) DeleteChirp(id int, subject int) (int, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return 500, err
	}
	if dbStructure.Chirps[id].Author != subject {
		return 403, errors.New("unauthorized")
	}
	log.Printf("DB: Attempting to delete chirp id %v with author %v", id, subject)
	delete(dbStructure.Chirps, id)
	db.mu.Lock()
	err = db.writeDB(dbStructure)
	db.mu.Unlock()

	if err != nil {
		return 500, err
	}

	return 200, nil
}
