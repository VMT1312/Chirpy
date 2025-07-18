package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/VMT1312/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database       *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w, r)
		},
	)
}

func (cfg *apiConfig) getFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(
		`<html>
  			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`,
		cfg.fileserverHits.Load(),
	)),
	)
}

func (cfg *apiConfig) resetServerHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Fileserver hits reset"))
}
