package api

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	serverHits int
}

func NewApiConfig() *apiConfig{
	return &apiConfig{}
}

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		cfg.serverHits++
		fmt.Printf("Current hits: %v\n", cfg.serverHits)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) ResetHits(w http.ResponseWriter, r *http.Request) {
	cfg.serverHits = 0
	w.WriteHeader(http.StatusOK)
	fmt.Println("Server hits reset")
	}

func (cfg *apiConfig) GetHits(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.serverHits)))
}