package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	auth "github.com/clinto-bean/golang-servers/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

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

	userid, err := cfg.validateToken(authorization)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid user id type")
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

func (cfg *apiConfig) validateToken(arg string) (int, error) {

	const prefix = "Bearer "
	if !strings.HasPrefix(arg, prefix) {
		log.Println("Invalid token format.")
		return -1, errors.New("bad token format")
	}

	tokenString := strings.TrimPrefix(arg, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error){
		return []byte(cfg.JWTSecret), nil
	})
	
	if err != nil {
		log.Printf("Error: %v\n", err.Error())
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		issuer := claims["iss"]
		if issuer == nil {
			return 0, errors.New("issuer not found in token")
		}
		if issuer != "chirpy" {
			return 0, errors.New("incorrect issuer")
		}
		
		subject, ok := claims["sub"].(string)
		if !ok {
			return 0, errors.New("subject claim is missing or not a string")
		}

		convertedSubject, err := strconv.Atoi(subject)
		
		if err != nil {
			log.Println("Unable to convert subject")
			return 0, err
		}

		return convertedSubject, nil

	} else {
		log.Print("NOT OK! Invalid token.")
		return 0, errors.New("token is invalid or expired")
	}
}