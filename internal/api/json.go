package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) ValidateChirp (w http.ResponseWriter, req *http.Request) {

	type params struct {
		Body string `json:"body"`
	}

	type returnValue struct {
		Valid bool `json:"valid"`
	}

	if req.Method != "POST" {
		log.Printf("Request must be POST.")
		w.WriteHeader(400)
		return
	}

	decoder := json.NewDecoder(req.Body)
	parameters := params{}
	err := decoder.Decode(&parameters)

	if err != nil {
		log.Printf("Error processing JSON data: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(parameters.Body) > 140 {
		log.Printf("Chirp must be 140 characters or less")
		w.WriteHeader(400)
		return
	}

	respBody := returnValue{
		Valid: true,
	}

	dat, err := json.Marshal(respBody)

	if err != nil {
		log.Printf("Error marshalling response: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)

}