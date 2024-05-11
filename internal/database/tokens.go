package database

import (
	"errors"
	"log"
)

func (db *DB) CreateToken(body string, id int) (Token, error) {
	db.mu.RLock()
	dbStructure, err := db.loadDB()
	db.mu.RUnlock()
	if err != nil {
		return Token{}, err
	}
	tk := Token{}
	tk.Body = body
	tk.ID = id
	dbStructure.Tokens[body] = tk
	db.mu.Lock()
	defer db.mu.Unlock()
	err = db.writeDB(dbStructure)
	if err != nil {
		return Token{}, err
	}
	log.Println("DB: Successfully created token (refresh)")
	return Token{
		Body: tk.Body,
		ID:   tk.ID,
	}, nil
}

func (db *DB) GetToken(body string) (Token, error) {
	db.mu.RLock()
	dbStructure, err := db.loadDB()
	db.mu.RUnlock()
	if err != nil {
		return Token{}, err
	}
	if t, ok := dbStructure.Tokens[body]; ok {
		return t, nil
	}
	return Token{}, errors.New("DB error: token does not exist")
}

func (db *DB) DeleteToken(body string) error {

	db.mu.RLock()
	dbStructure, err := db.loadDB()
	db.mu.RUnlock()
	if err != nil {
		return err
	}
	if _, ok := dbStructure.Tokens[body]; ok {
		log.Println("DB: refresh token found. deleting.")
		delete(dbStructure.Tokens, body)
		db.mu.Lock()
		defer db.mu.Unlock()
		err = db.writeDB(dbStructure)
		if err != nil {
			return err
		}
		log.Println("DB: Successfully deleted refresh token")
		return nil
	}
	log.Println("DB: Refresh token was not found")
	return errors.New("DB error: resource not found")
}
