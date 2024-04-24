package main

import (
	"log"
	"net/http"
	"os"

	db "github.com/clinto-bean/golang-servers/internal/database"
	godotenv "github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	DB             *db.DB
	JWTSecret string
	Expiration int
}

func main() {
	const root = "./"
	const port = "8080"
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := db.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
		JWTSecret: jwtSecret,
		Expiration: 5,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", apiCfg.handlerReadiness)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetSingleChirp)
	mux.HandleFunc("GET /api/users/", apiCfg.handlerGetAllUsers)
	mux.HandleFunc("GET /api/users/{userID}", apiCfg.handlerGetSingleUser)
	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUsers)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	corsMux := middlewareCors(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Server running on port %v", port)
	log.Fatal(srv.ListenAndServe())
}