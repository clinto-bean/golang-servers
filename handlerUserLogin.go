package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/clinto-bean/golang-servers/internal/auth"
)

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {

	log.Print("Attempting to log in")

	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Print("Could not decode parameters")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	dbUser, err := cfg.DB.GetUserByEmail(params.Email)
	
	if err != nil {
		log.Print("User not found")
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	err = auth.CheckPasswords(params.Password, dbUser.Password)

	if err != nil {
		log.Print("Passwords do not match!")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	
	}

	log.Printf("Successfully logged in as %v!\n", params.Email)

	respondWithJSON(w, http.StatusOK, User{
		ID: dbUser.ID,
		Email: dbUser.Email,
	})

}