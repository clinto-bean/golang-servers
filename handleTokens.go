package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type returnParams struct {
		Token string `json:"token"`
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		log.Println("No token found")
		respondWithError(w, http.StatusBadRequest, "no token found")
		return
	}

	userid, err := cfg.validateToken(auth, "chirpy-refresh")

	if err != nil {
		log.Print("did not validate")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = cfg.DB.GetToken(strings.TrimPrefix(auth, "Bearer "))

	if err != nil {
		log.Print("could not verify that the db token exists. access denied")
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := cfg.generateUserToken(userid, time.Now().Add(time.Hour), "chirpy-access")

	if err != nil {
		log.Println("couldn't generate token")
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("API: Refreshed! New access token created")

	respondWithJSON(w, http.StatusOK, returnParams{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "
	if strings.HasPrefix(auth, prefix) {
		auth = strings.TrimPrefix(auth, prefix)
	} else {
		respondWithError(w, http.StatusBadRequest, "token has incorrect format")
		return
	}
	log.Println("API: Attempting to delete token (refresh)")
	err := cfg.DB.DeleteToken(auth)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	log.Println("API: Successfully revoked refresh token")
	respondWithJSON(w, http.StatusOK, nil)
}

func (cfg *apiConfig) generateUserToken(userid int, expiration time.Time, issuer string) (string, error) {

	now := jwt.NewNumericDate(time.Now())
	exp := jwt.NewNumericDate(expiration)
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  now,
		ExpiresAt: exp,
		Subject:   strconv.Itoa(userid),
	})
	t, err := tkn.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		log.Print(err)
		return "", err
	}
	return t, nil
}

func (cfg *apiConfig) validateToken(arg string, issuerName string) (int, error) {

	const prefix = "Bearer "
	if !strings.HasPrefix(arg, prefix) {
		log.Println("Invalid token format.")
		return -1, errors.New("bad token format")
	}

	tokenString := strings.TrimPrefix(arg, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		log.Printf("Error: %v\n", err.Error())
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		issuer := claims["iss"]
		if issuer != issuerName {

			return 0, errors.New("incorrect issuer")
		}

		if issuer == nil {
			return 0, errors.New("issuer not found in token")
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
