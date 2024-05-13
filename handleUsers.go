package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/clinto-bean/golang-servers/internal/auth"
)

type User struct {
	Email   string `json:"email"`
	Premium bool   `json:"is_chirpy_red"`
	ID      int    `json:"id"`
}

// handleCreateUsers attempts to create the user entry in the database and notifies requester of any issues processing

func (cfg *apiConfig) handleCreateUsers(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

	// 1: decode request object for username and password

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	email := params.Email
	password := params.Password
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2: validate that the email address provided matches required format

	e, err := validateEmail(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3: encrypt user password

	p, err := auth.EncryptPassword(password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 4: create database entry for user

	user, err := cfg.DB.CreateUser(e, p, false)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	// 5: successfully respond with the newly created user object

	respondWithJSON(w, http.StatusCreated, User{
		Email:   user.Email,
		Premium: user.Premium,
		ID:      user.ID,
	})

}

func validateEmail(email string) (string, error) {

	// 1: validate whether email meets required format, return empty string and error if it does not, or simply the email and nil error if it does

	if !strings.Contains(email, "@") {
		return "", errors.New("please enter a valid email")
	}
	return email, nil
}

// handlerGetAllUsers loads all users from the database then iterates over them, creating a new slice of users and returning it

func (cfg *apiConfig) handlerGetAllUsers(w http.ResponseWriter, r *http.Request) {

	// 1: attempt to get users from ddatabase

	dbUsers, err := cfg.DB.GetUsers()
	if err != nil {
		fmt.Print("unable to run cfg.DB.GetUsers()")
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	// 2: initialize new slice of users

	users := []User{}

	// 3: iterate over dbUsers and append each user to the users slice

	for _, user := range dbUsers {
		users = append(users, User{
			ID:      user.ID,
			Premium: user.Premium,
			Email:   user.Email,
		})
	}

	// 4: sort users by ascending id then return the list of users

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	respondWithJSON(w, http.StatusOK, users)
	
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
		Email:   user.Email,
		ID:      user.ID,
		Premium: user.Premium,
	})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		ID       string `json:"id"`
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

	u, err := cfg.DB.UpdateUser(userid, params.Email, pw, false)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error occurred while updating user")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		Email:   u.Email,
		Premium: u.Premium,
		ID:      u.ID,
	})

}
