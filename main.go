package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/horsedevours/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      database.Queries
	platform       string
}

func main() {
	godotenv.Load()

	dbString := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbString)
	if err != nil {
		fmt.Printf("error connecting to database: %v", err)
		os.Exit(1)
	}

	cfg := &apiConfig{
		dbQueries: *database.New(db),
		platform:  os.Getenv("PLATFORM"),
	}

	smx := http.NewServeMux()
	smx.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	smx.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	smx.HandleFunc("POST /admin/reset", cfg.handlerReset)
	smx.HandleFunc("GET /admin/healthz", handlerHealthz)
	smx.HandleFunc("POST /api/users", cfg.handlerPostUser)
	smx.HandleFunc("POST /api/chirps", cfg.handlerPostChirp)
	smx.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	smx.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirp)

	srvr := http.Server{
		Handler: smx,
		Addr:    ":8080",
	}

	srvr.ListenAndServe()
}
