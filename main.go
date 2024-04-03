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
	router.Handle("/app/*", api.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root)))))
	router.HandleFunc("/healthz", handleReadiness)
	router.HandleFunc("/metrics", api.GetHits)
	router.HandleFunc("/reset", api.ResetHits)

	corsMux := middlewareCors(router)

	log.Printf("Listening on port %s", port)

	s := &http.Server{
		Addr: ":" + port,
		Handler: corsMux,
	}

	log.Fatal(s.ListenAndServe())
	
}

func handleReadiness(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
}