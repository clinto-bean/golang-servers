package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/clinto-bean/golang-servers/internal/auth"
)

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {

	log.Println("API: Logging in")

	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		JWT              string `json:"jwt,omitempty"`
		ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty"`
	}

	type returnParams struct {
		ID      int    `json:"id"`
		Email   string `json:"email"`
		Token   string `json:"token"`
		Refresh string `json:"refresh_token"`
		Premium bool   `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	log.Println("API: Decoding login parameters")

	if err != nil {
		log.Print("Could not decode parameters")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	log.Print("API: Attempting to get user from Database")

	dbUser, err := cfg.DB.GetUserByEmail(params.Email)

	if err != nil {
		log.Print("User not found")
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	log.Print("API: Attempting to validate password")

	err = auth.CheckPasswords(params.Password, dbUser.Password)

	if err != nil {
		log.Print("Passwords do not match!")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return

	}

	log.Println("API: Attempting to create access token")

	now := time.Now()

	token, err := cfg.generateUserToken(dbUser.ID, now.Add(time.Hour), "chirpy-access")

	log.Println("API: Generated token (access)")

	if err != nil {
		log.Print("Unable to generate access token")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("API: Attempting to create refresh token")

	refresh, err := cfg.generateUserToken(dbUser.ID, now.Add(time.Hour*24*60), "chirpy-refresh")

	if err != nil {
		log.Println("Unable to generate refresh token")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = cfg.DB.CreateToken(refresh, dbUser.ID)

	if err != nil {
		log.Println("Could not save refresh token to db")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Println("API: Token generated (refresh)")

	respondWithJSON(w, http.StatusOK, returnParams{
		ID:      dbUser.ID,
		Email:   dbUser.Email,
		Token:   token,
		Refresh: refresh,
		Premium: dbUser.Premium,
	})
}
