package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/VMT1312/Chirpy/internal/auth"
	"github.com/VMT1312/Chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	JWTSecret      string
	polkaKey       string
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

func (cfg *apiConfig) resetUserTable(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "This endpoint is only available in development mode")
		return
	}

	if err := cfg.db.ResetUser(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reset user table")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{"message": "User table reset successfully"})
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := parameter{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	arg := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
	}
	dbUser, err := cfg.db.CreateUser(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	user := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.ChirpyRed,
	}

	respondWithJson(w, http.StatusCreated, user)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token: "+err.Error())
		return
	}

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
	} else {
		respondWithError(w, http.StatusBadRequest, "Body exceeds 140 characters")
	}

	arg := database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	}

	dbChirp, err := cfg.db.CreateChirp(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID.String(),
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.String(),
	}
	respondWithJson(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps")
		return
	}

	chirps := make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		chirps[i] = Chirp{
			ID:        dbChirp.ID.String(),
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID.String(),
		}
	}

	respondWithJson(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	stringID := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(stringID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirp")
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID.String(),
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.String(),
	}

	respondWithJson(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	params := parameter{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	err = auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create JWT token")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	arg := database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: dbUser.ID,
	}

	refreshTokenDB, err := cfg.db.CreateRefreshToken(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token in database")
		return
	}

	user := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refreshTokenDB.Token,
		IsChirpyRed:  dbUser.ChirpyRed,
	}

	respondWithJson(w, http.StatusOK, user)
}

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	dbToken, err := cfg.db.GetRefreshTokenByToken(r.Context(), bearerToken)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve refresh token")
		return
	}

	if time.Now().After(dbToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired")
		return
	} else if dbToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked")
		return
	}

	token, err := auth.MakeJWT(dbToken.UserID, cfg.JWTSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create JWT token")
		return
	}

	resToken := AccessToken{
		Token: token,
	}

	respondWithJson(w, http.StatusOK, resToken)
}

func (cfg *apiConfig) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), bearerToken)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token: "+err.Error())
		return
	}

	params := parameter{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	arg := database.UpdatePasswordByIDParams{
		HashedPassword: hashedPassword,
		ID:             userID,
		Email:          params.Email,
	}
	dbUser, err := cfg.db.UpdatePasswordByID(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	user := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.ChirpyRed,
	}
	respondWithJson(w, http.StatusOK, user)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp ID is required")
		return
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token: "+err.Error())
		return
	}

	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format")
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirp")
		return
	}
	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You do not have permission to delete this chirp")
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) upgradeUserHandler(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusForbidden, "Invalid API key")
		return
	}

	decoder := json.NewDecoder(r.Body)

	params := parameter{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.db.UpgradeUserByID(r.Context(), params.Data.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
