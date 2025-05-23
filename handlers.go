package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/horsedevours/chirpy/internal/database"
)

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	writeHtml(w, http.StatusOK, "OK")
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	html := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>	
	`
	writeHtml(w, http.StatusOK, fmt.Sprintf(html, int(cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.dbQueries.DeleteAllUsers(context.Background())
	cfg.fileserverHits.Swap(int32(0))

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerPostUser(w http.ResponseWriter, r *http.Request) {
	email := struct {
		Email string `json:"email"`
	}{}
	err := unmarshalRequestBody(r.Body, &email)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}

	user, err := cfg.dbQueries.CreateUser(context.Background(), email.Email)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}

	resp, err := json.Marshal(user)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}
	writeJson(w, http.StatusCreated, resp)
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	chirp := struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}{}

	err := unmarshalRequestBody(r.Body, &chirp)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
	}

	if len(chirp.Body) > 140 {
		err = errors.New("Chirp is too long")
		writeErrorResponse(w, err, http.StatusBadRequest, err.Error())
		return
	}

	badWords := map[string]struct{}{"kerfuffle": {}, "sharbert": {}, "fornax": {}}
	chirpWords := strings.Split(chirp.Body, " ")

	for i, word := range chirpWords {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			chirpWords[i] = "****"
		}
	}

	cleanedWords := strings.Join(chirpWords, " ")

	savedChirp, err := cfg.dbQueries.CreateChirp(context.Background(), database.CreateChirpParams{
		Body:   cleanedWords,
		UserID: uuid.MustParse(chirp.UserId),
	})
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}

	resp, err := json.Marshal(savedChirp)
	if err != nil {
		log.Printf("error: %v", err)
	}
	writeJson(w, http.StatusCreated, resp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetAllChirps(context.Background())
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}

	data, err := json.Marshal(chirps)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}
	writeJson(w, http.StatusOK, data)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	chirp, err := cfg.dbQueries.GetChirpById(context.Background(), uuid.MustParse(id))
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError, INTERNAL_SERVER_MESSAGE)
		return
	}
	writeJson(w, http.StatusOK, data)
}
