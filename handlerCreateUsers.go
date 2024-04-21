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

	if err != nil {
		
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	e, err := validateEmail(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	p, err := auth.EncryptPassword(password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

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
	return email, nil
}