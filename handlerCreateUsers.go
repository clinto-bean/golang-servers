package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/clinto-bean/golang-servers/internal/auth"
)



type User struct {
	Email string `json:"email"`
	ID int `json:"id"`
}

type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
}

func (cfg *apiConfig) handleCreateUsers(w http.ResponseWriter, r *http.Request) {
	

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	email := params.Email
	password := params.Password
	log.Printf("Email: %v, password: %v\n", email, password)
	if err != nil {
		
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("Attempting to validate email")

	e, err := validateEmail(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("Requesting password encryption")

	p, err := auth.EncryptPassword(password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("Attempting to create user in database")

	user, err := cfg.DB.CreateUser(e, p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	

	respondWithJSON(w, http.StatusCreated, User{
		Email: user.Email,
		ID: user.ID,
	})

}

func validateEmail(email string) (string, error) {
	if !strings.Contains(email, "@") {
		log.Println("Invalid email")
		return "", errors.New("please enter a valid email")
	}
	log.Println("Email validated")
	return email, nil
}