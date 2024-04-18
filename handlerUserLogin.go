package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/clinto-bean/golang-servers/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {

	log.Print("Attempting to log in")

	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
		JWT string `json:"jwt,omitempty"`
		ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty"`
	}
	
	type returnParams struct {
		Token string `json:"token"`
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

	exp := cfg.Expiration

	if params.ExpiresInSeconds != nil && *params.ExpiresInSeconds < int64(24*time.Hour.Seconds()) {
		exp = int(*params.ExpiresInSeconds)
	}
	
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

	log.Println("Attempting to create user token")

	token, err := cfg.generateUserToken(dbUser.ID, exp)

	if err != nil {
		log.Print("Unable to generate user token")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)

	respondWithJSON(w, http.StatusOK, returnParams{
		Token: token,
	})
}

func (cfg *apiConfig) generateUserToken(userid int, expiration int) (string, error) {

	now := jwt.NewNumericDate(time.Now())
	exp := jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expiration)))
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: now,
		ExpiresAt: exp,
		Subject: strconv.Itoa(userid),
	})
	t, err := tkn.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		log.Print(err)
		return "", err
	}
	
	return t, nil
}