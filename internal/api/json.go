package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (cfg *apiConfig) ValidateChirp (w http.ResponseWriter, req *http.Request) {

	const maxChirpChars = 140

	type params struct {
		Body string `json:"body"`
	}

	if req.Method != "POST" {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Method must be POST. Current method: %s", req.Method))
		return
	}

	decoder := json.NewDecoder(req.Body)
	parameters := params{}
	err := decoder.Decode(&parameters)

	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling request: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	if len(parameters.Body) > maxChirpChars {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Message cannot be longer than %v characters.", maxChirpChars))
		return
	}

	filterJSON(w, 200, strings.Split(parameters.Body, " "))

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Server error 5xx: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling response: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func filterJSON(w http.ResponseWriter, code int, words []string) {

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	var filteredChirp []string
	for _, word := range words {
		profanity := []string{"kerfuffle", "sharbert", "fornax"}
		for i := range profanity {
			if strings.EqualFold(word, profanity[i]) {
				word = "****"
			}
		}
		filteredChirp = append(filteredChirp, word)
	}

	msg := returnVals{CleanedBody: strings.Join(filteredChirp, " ")}

	respondWithJSON(w, code, msg)
	
}