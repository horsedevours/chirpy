package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/horsedevours/chirpy/internal/database"
	_ "github.com/jackc/pgx"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbString := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbString)
	if err != nil {
		fmt.Printf("error connecting to database: %v", err)
		os.Exit(1)
	}
	cfg := &apiConfig{dbQueries: *database.New(db)}
	smx := http.NewServeMux()
	smx.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	smx.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	smx.HandleFunc("POST /admin/reset", cfg.handlerReset)
	smx.HandleFunc("GET /admin/healthz", handlerHealthz)
	smx.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	smx.HandleFunc("POST /api/users", handlerPostUser)

	srvr := http.Server{
		Handler: smx,
		Addr:    ":8080",
	}

	srvr.ListenAndServe()
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	chirp := struct {
		Body string `json:"body"`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		w.WriteHeader(500)
		w.Write([]byte("{\"error\":\"Something went wrong\"}"))
		return
	}

	if len(chirp.Body) > 140 {
		log.Printf("Chirp is too long: %s", chirp.Body)
		w.WriteHeader(400)
		w.Write([]byte("{\"error\":\"Chirp is too long\"}"))
		return
	}

	badWords := map[string]struct{}{"kerfuffle": struct{}{}, "sharbert": struct{}{}, "fornax": struct{}{}}
	chirpWords := strings.Split(chirp.Body, " ")

	for i, word := range chirpWords {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			chirpWords[i] = "****"
		}
	}

	cleanedWords := strings.Join(chirpWords, " ")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("{\"cleaned_body\":\"%s\"}", cleanedWords)))
	return
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(int32(1))
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	html := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>	
	`
	w.Write([]byte(fmt.Sprintf(html, int(cfg.fileserverHits.Load()))))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Swap(int32(0))
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerPostUser(w http.ResponseWriter, r *http.Request) {
	email := struct {
		Email string `jsont:"email"`
	}{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %v", err)
		return
	}
	err = json.Unmarshal(data, &email)
	if err != nil {
		log.Printf("error unmarshalling request data: %v", err)
		return
	}

	user, err := cfg.dbQueries.CreateUser(context.Background(), sql.NullString{String: email.Email, Valid: true})
	if err != nil {
		log.Printf("error creating user: %v", err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp, err := json.Marshal(user)
	if err != nil {
		log.Printf("error marshalling user: %v", err)
		return
	}
	w.Write(resp)
	return
}
