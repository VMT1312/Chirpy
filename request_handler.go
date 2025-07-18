package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type parameter struct {
	Body string `json:"body"`
}

type successRes struct {
	Clean_body string `json:"cleaned_body"`
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := parameter{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if len(params.Body) <= 140 {
		words := strings.Split(params.Body, " ")
		for i, word := range words {
			word = strings.ToLower(word)
			if _, ok := bannedWords[word]; ok {
				words[i] = "****"
			}
		}

		res := successRes{
			Clean_body: strings.Join(words, " "),
		}
		respondWithJson(w, http.StatusOK, res)
		return
	} else {
		respondWithError(w, http.StatusBadRequest, "Body exceeds 140 characters")
	}
}
