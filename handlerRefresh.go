package main

import (
	"log"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type returnParams struct {
		Token string `json:"token"`
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		log.Println("No token found")
		respondWithError(w, http.StatusBadRequest, "no token found")
		return
	}

	userid, err := cfg.validateToken(auth, "chirpy-refresh")
	if err != nil {
		log.Println(err.Error())
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := cfg.generateUserToken(userid, time.Now().Add(time.Hour), "chirpy-access")
	if err != nil {
		log.Println("could not generate new token")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, returnParams{
		Token: token,
	})
}