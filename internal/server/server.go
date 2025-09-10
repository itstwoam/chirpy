package server

import (
	"net/http"
	"fmt"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(1)	
		next.ServeHTTP(w, r) 
	})
}

func (cfg *apiConfig) serveHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
	w.Write([]byte("Hits: "+ strconv.Itoa(int(hits))))
}

func (cfg *apiConfig) serveReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func StartServer() {
	cfg := &apiConfig{}
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./internal/app"))
	//mux.Handle("/app/", http.StripPrefix("/app", fs))
	mux.Handle("/app/", cfg.middlewareInc(http.StripPrefix("/app", fs)))
	mux.HandleFunc("GET /healthz", serveStatus)
	mux.HandleFunc("GET /metrics", cfg.serveHits)
	mux.HandleFunc("POST /reset", cfg.serveReset)
	server := http.Server{}
	server.Handler = mux
	server.Addr = ":8085"
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("There was an error, but wtf?")
	}
	fmt.Println("Am I still running?")
}

func serveStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
