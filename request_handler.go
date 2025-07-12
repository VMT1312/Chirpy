package main

import (
	"encoding/json"
	"net/http"
)

type parameter struct {
	Body string `json:"body"`
}

type successRes struct {
	Valid bool `json:"valid"`
}

type failRes struct {
	Error string `json:"error"`
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := parameter{}
	if err := decoder.Decode(&params); err != nil {
		res := failRes{Error: "Something went wrong"}
		if dat, err := json.Marshal(res); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(dat)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(params.Body) <= 140 {
		res := successRes{Valid: true}
		if dat, err := json.Marshal(res); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(dat)
		}
	} else {
		res := failRes{Error: "Chirp is too long"}
		if dat, err := json.Marshal(res); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(dat)
		}
	}

}
