package main

import (
	"log"
	"net/http"

	"github.com/clinto-bean/golang-servers/internal/api"
)

func main () {
	const root = "./"
	const port = "8080"

	api := api.NewApiConfig()

	router := http.NewServeMux()
	
	handleFileSystem := api.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	
	router.Handle("GET /app/*", handleFileSystem)
	router.HandleFunc("GET /api/healthz", api.HandleReadiness)
	router.HandleFunc("GET /admin/metrics", api.GetHits)
	router.HandleFunc("GET /api/reset", api.ResetHits)

	corsMux := middlewareCors(router)

	log.Printf("Listening on port %s", port)

	s := &http.Server{
		Addr: ":" + port,

		Handler: corsMux,
	}

	log.Fatal(s.ListenAndServe())
	
}

