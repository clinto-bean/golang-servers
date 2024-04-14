package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		fmt.Print("unable to run cfg.DB.GetChirps()")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:   dbChirp.ID,
			Body: dbChirp.Body,
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
		fmt.Print("unable to convert chirp ID")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chirp, err := cfg.DB.GetChirp(id)
	if err != nil {
		fmt.Print("unable to get chirp by id")
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
				Body: chirp.Body,
				ID: chirp.ID,
			})
}