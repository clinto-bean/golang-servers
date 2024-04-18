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

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		log.Print("Could not load db.")
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
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

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users: map[int]User{},
	}

	db.mu.Lock()
	err := db.writeDB(dbStructure)
	db.mu.Unlock()

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
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
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

	db.mu.Lock()
	err = db.writeDB(dbStructure)
	db.mu.Unlock()

	if err != nil {
		log.Print("Couldn't write user to db")
		return User{}, err
	}
	
	return user, nil
}

func (db *DB) UpdateUser(id int, email string, password string) (User, error) {
	dbStructure, err := db.loadDB()
	db.mu.Lock()
	defer db.mu.Unlock()

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