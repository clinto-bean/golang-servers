package database

import (
	"errors"
	"log"
)

func (db *DB) CreateUser(email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		log.Print("Couldn't load db while creating users")
		return User{}, err
	}

	for i := range dbStructure.Users {
		if dbStructure.Users[i].Email == email {
			return User{}, errors.New("user already exists")
		}
	}

	id := len(dbStructure.Users) + 1

	user := User{
		Email: email,
		ID: id,
		Password: password,
	}

	dbStructure.Users[id] = user

	db.mu.RLock()
	err = db.writeDB(dbStructure)
	db.mu.RUnlock()

	if err != nil {
		log.Print("Couldn't write user to db")
		return User{}, err
	}
	
	return user, nil
}

func (db *DB) UpdateUser(id int, email string, password string) (User, error) {

	db.mu.RLock()
	dbStructure, err := db.loadDB()
	db.mu.RUnlock()

	if err != nil {
		log.Print("Couldn't load db while updating users")
		return User{}, err
	}

	user, ok := dbStructure.Users[id]

	if !ok {
		return User{}, errors.New("user not found")
	}

	user.Email = email
	user.Password = password
	user.ID = id

	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	

	if err != nil {
		return User{}, errors.New("could not write database")
	}

	return user, nil
}

func (db *DB) GetUsers() ([]User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		log.Print("Could not get chirps.")
		return nil, err
	}

	users := make([]User, 0, len(dbStructure.Users))
	for _, user := range dbStructure.Users {
		users = append(users, user)
	}

	return users, nil
}

func (db *DB) GetSingleUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
		if err != nil {
			return User{}, err
		}
		user, ok := dbStructure.Users[id]
		if !ok {
			return User{}, ErrNotExist
		}
		return user, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
		if err != nil {
			return User{}, err
		}
		for _, u := range dbStructure.Users {
			if u.Email == email {
				return u, nil
			}
		}
		return User{}, ErrNotExist
}