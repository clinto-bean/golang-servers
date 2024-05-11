package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Chirp struct {
	ID     int    `json:"id"`
	Body   string `json:"body"`
	Author int    `json:"author_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters while creating chirp")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	token := r.Header.Get("Authorization")
	subject, err := cfg.validateToken(token, "chirpy-access")
	if err != nil {
		respondWithError(w, 500, "could not determine access token")
		return
	}

	chirp, err := cfg.DB.CreateChirp(cleaned, subject)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:     chirp.ID,
		Body:   chirp.Body,
		Author: chirp.Author,
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		log.Println("unable to get chirps")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:     dbChirp.ID,
			Body:   dbChirp.Body,
			Author: dbChirp.Author,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		log.Println("unable to convert chirp ID")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chirp, err := cfg.DB.GetChirp(id)
	if err != nil {
		log.Println("unable to get chirp by id")
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		Body:   chirp.Body,
		ID:     chirp.ID,
		Author: chirp.Author,
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		log.Println("API: Unable to convert chirp ID")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	token := r.Header.Get("Authorization")
	subject, err := cfg.validateToken(token, "chirpy-access")

	if err != nil {
		log.Println("API: Could not validate token in DeleteChirp")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	status, err := cfg.DB.DeleteChirp(id, subject)
	if err != nil {
		log.Println("API: Could not delete chirp")
		respondWithError(w, status, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, "")
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
