package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerGetAllUsers(w http.ResponseWriter, r *http.Request) {
	dbUsers, err := cfg.DB.GetUsers()
	if err != nil {
		fmt.Print("unable to run cfg.DB.GetUsers()")
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	users := []User{}

	for _, user := range dbUsers {
		users = append(users, User{
			ID: user.ID,
			Email: user.Email,
		})
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	respondWithJSON(w, http.StatusOK, users)
	fmt.Print(users)
}

func (cfg *apiConfig) handlerGetSingleUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	id, err := strconv.Atoi(userID)
	if err != nil && userID != "" {
		fmt.Print("could not convert user ID")
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	user, err := cfg.DB.GetSingleUser(id)
	if err != nil {
		fmt.Print("unable to locate user by id")
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, User{
		Email: user.Email,
		ID: user.ID,
	})
}