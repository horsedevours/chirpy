package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

func main() {
	cfg := &apiConfig{}
	smx := http.NewServeMux()
	smx.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	smx.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	smx.HandleFunc("POST /admin/reset", cfg.handlerReset)
	smx.HandleFunc("GET /api/healthz", handlerHealthz)
	smx.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

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
