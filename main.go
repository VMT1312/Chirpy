package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

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
