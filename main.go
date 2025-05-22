package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

func main() {
	cfg := &apiConfig{}
	smx := http.NewServeMux()
	smx.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	smx.HandleFunc("/healthz", handlerHealthz)
	smx.HandleFunc("/metrics", cfg.handlerMetrics)
	smx.HandleFunc("/reset", cfg.handlerReset)

	srvr := http.Server{
		Handler: smx,
		Addr:    ":8080",
	}

	srvr.ListenAndServe()
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
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
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %s", strconv.Itoa(int(cfg.fileserverHits.Load())))))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Swap(int32(0))
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
