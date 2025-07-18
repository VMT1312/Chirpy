package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/VMT1312/Chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		database: dbQueries,
	}

	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))))

	mux.HandleFunc("GET /api/healthz", healthCheckHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getFileserverHits)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetServerHits)

	mux.HandleFunc("POST /api/validate_chirp", requestHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
