package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type User struct {
	Email string `json:"email"`
	ID int `json:"id"`
}

func (cfg *apiConfig) handleCreateUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	email := params.Body
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	data, err := validateUser(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	user, err := cfg.DB.CreateUser(data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusCreated, User{
		Email: user.Email,
		ID: user.ID,
	})

}

func validateUser(email string) (string, error) {
	if !strings.Contains(email, "@") {
		return "", errors.New("please enter a valid email")
	}
	return email, nil
}