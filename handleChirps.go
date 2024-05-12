package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

/* 	handlerChirpsCreate creates a chirp, saves it to database and sends it back via response */

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	// 1: attempt to decode json data from request object

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters while creating chirp")
		return
	}

	// 2: pass chirp to validator function to ensure it meets necessary standards

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3: parse and validate user's access token

	token := r.Header.Get("Authorization")
	subject, err := cfg.validateToken(token, "chirpy-access")
	if err != nil {
		respondWithError(w, 500, "could not determine access token")
		return
	}

	// 4: if access token is valid, create chirp in database

	chirp, err := cfg.DB.CreateChirp(cleaned, subject)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	// 5: respond successfully with copy of created chirp

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:     chirp.ID,
		Body:   chirp.Body,
		Author: chirp.Author,
	})
}

/* handlerGetAllChirps receives author_id as a parameter and returns all chirps for that user */

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		AuthorID int `json:"author_id"`
	}

	// 1: attempt to parse author id from provided request

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters while fetching all chirps")
		return
	}

	// 2: attempt to retrieve all chirps from database

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		log.Println("unable to get chirps")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3: filter chirps into new slice which match the author_id

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		if dbChirp.Author == params.AuthorID {
			chirps = append(chirps, Chirp{
			ID:     dbChirp.ID,
			Body:   dbChirp.Body,
			Author: dbChirp.Author,
		})}
	}

	// 4: if no matching chirps, successfully respond stating no chirps found

	if len(chirps) < 1 {
		respondWithJSON(w, http.StatusOK, fmt.Sprintf("No chirps found for author $%v", params.AuthorID))
		return
	}

	// 5: sort chirps by ID if matches were found

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	// 6: respond successfully with requested list of chirps

	respondWithJSON(w, http.StatusOK, chirps)
}

/* handlerGetSingleChirp returns a chirp based on its database ID */

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {

	// 1: attempt to parse chirp ID from request object provided via url params

	chirpID := r.PathValue("chirpID")
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		log.Println("unable to convert chirp ID")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2: attempt to locate chirp in db, respond with 404 and return

	chirp, err := cfg.DB.GetChirp(id)
	if err != nil {
		log.Println("unable to get chirp by id")
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	// 3: successfully respond with requested chirp

	respondWithJSON(w, http.StatusOK, Chirp{
		Body:   chirp.Body,
		ID:     chirp.ID,
		Author: chirp.Author,
	})
}

/* handlerDeleteChirp parses chirp ID from url parameters and attempts to delete from database */

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {

	// 1: attempt to parse chirp ID from request object via url parameters

	chirpID := r.PathValue("chirpID")
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		log.Println("API: Unable to convert chirp ID")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2: verify and validate user's access token

	token := r.Header.Get("Authorization")
	subject, err := cfg.validateToken(token, "chirpy-access")
	if err != nil {
		log.Println("API: Could not validate token in DeleteChirp")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3: attempt to create the chirp in the database, if successful, return it, if not, return error

	status, err := cfg.DB.DeleteChirp(id, subject)
	if err != nil {
		log.Println("API: Could not delete chirp")
		respondWithError(w, status, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, "")
}

/* validateChirp receives the chirp body and ensures its length is below 140 characters
any chirps exceeding 140 characters will be rejected with error "Chirp is too long"
pre-defined profane words will be replaced before the chirp is added to the database */

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140

	// 1: ensure the length of the body does not exceed the max chirp length, return an error if it does

	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	// 2: define words to be filtered out of message

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	// 3: pass the body and badWords to getCleanedBody then return the result

	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

// getCleanedBody iterates over each word in the chirp and replaces matching text with asterisks

func getCleanedBody(body string, badWords map[string]struct{}) string {

	// 1: separate the string delimited with whitespace

	words := strings.Split(body, " ")

	// 2: iterate over words and replace text matching badWords pattern

	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}

	// 3: respond with the mutated text in the form of a string

	cleaned := strings.Join(words, " ")
	return cleaned
}
