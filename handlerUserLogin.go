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

	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
		JWT string `json:"jwt,omitempty"`
		ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty"`
	}
	
	type returnParams struct {
		ID int `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
		Refresh string `json:"refresh_token"`
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

	now := time.Now()

	token, err := cfg.generateUserToken(dbUser.ID, now.Add(time.Hour), "chirpy-access")

	if err != nil {
		log.Print("Unable to generate access token")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refresh, err := cfg.generateUserToken(dbUser.ID, now.Add(time.Hour * 24 * 60), "chirpy-refresh")

	if err != nil {
		log.Print("Unable to generate refresh token")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println()
	log.Println("{")
	log.Printf("token: %v\n", token)
	log.Printf("refresh: %v\n", refresh)
	log.Println("}")
	
	respondWithJSON(w, http.StatusOK, returnParams{
		ID: dbUser.ID,
		Email: dbUser.Email,
		Token: token,
		Refresh: refresh,
	})
}


func (cfg *apiConfig) generateUserToken(userid int, expiration time.Time, issuer string) (string, error) {

	now := jwt.NewNumericDate(time.Now())
	exp := jwt.NewNumericDate(expiration)
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: issuer,
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
