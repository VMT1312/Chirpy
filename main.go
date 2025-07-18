package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/VMT1312/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type parameter struct {
	Body  string `json:"body"`
	Email string `json:"email"`
}

func main() {
	godotenv.Load(".env")

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		db:       dbQueries,
		platform: os.Getenv("PLATFORM"),
	}

	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))))

	mux.HandleFunc("GET /api/healthz", healthCheckHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getFileserverHits)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetUserTable)

	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)

	mux.HandleFunc("POST /api/validate_chirp", requestHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
