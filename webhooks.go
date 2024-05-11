package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type webhookParams struct {
	Event string         `json:"event"`
	Data  map[string]int `json:"data"`
}

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := webhookParams{}
	err := decoder.Decode(&params)

	auth := r.Header.Get("Authorization")
	reqApiKey := strings.TrimPrefix(auth, "ApiKey ")

	if reqApiKey != cfg.APIKey {
		respondWithError(w, http.StatusUnauthorized, "API: API Key is invalid")
		return
	}

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "API: Could not decode json in webhook.")
		return
	}

	event := params.Event
	data := params.Data

	if event == "user.payment_failed" {
		respondWithError(w, http.StatusOK, "API: Payment failed")
		return
	}

	if event == "user.upgraded" {
		user := data["user_id"]

		dbUser, err := cfg.DB.GetSingleUser(user)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "API: could not fetch DBUser in handlerUpgradeUser")
			return
		}

		if !dbUser.Premium {
			dbUser.Premium = true
			cfg.DB.UpdateUser(user, dbUser.Email, dbUser.Password, dbUser.Premium)
		}

		respondWithJSON(w, http.StatusOK, nil)

		return
	}

	respondWithError(w, http.StatusInternalServerError, "API: Invalid request to handlerUpgradeUser")

}
