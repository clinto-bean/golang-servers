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

	"github.com/clinto-bean/golang-servers/internal/auth"
)



type User struct {
	Email string `json:"email"`
	ID int `json:"id"`
}

type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
}

func (cfg *apiConfig) handleCreateUsers(w http.ResponseWriter, r *http.Request) {
	

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	email := params.Email
	password := params.Password

	if err != nil {
		
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	e, err := validateEmail(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	p, err := auth.EncryptPassword(password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := cfg.DB.CreateUser(e, p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}


	respondWithJSON(w, http.StatusCreated, User{
		Email: user.Email,
		ID: user.ID,
	})

}

func validateEmail(email string) (string, error) {
	if !strings.Contains(email, "@") {
		log.Println("Invalid email")
		return "", errors.New("please enter a valid email")
	}
	return email, nil
}

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

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ID string `json:"id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not decode parameters")
		return
	}

	authorization := r.Header.Get("Authorization")

	userid, err := cfg.validateToken(authorization, "chirpy-access")

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	pw, err := auth.EncryptPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not encrypt new password")
		return
	}

	u, err := cfg.DB.UpdateUser(userid, params.Email, pw)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error occurred while updating user")
		return
	}

	
	respondWithJSON(w, http.StatusOK, User{
		Email: u.Email,
		ID: u.ID,
	})

}