package server

import (
	"net/http"
	"fmt"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareInc(next http.Handler) http.Handler {
	cfg.filserverHits.Add(1)	
}

func serveHits(w http.ResponseWrite, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(strconv.Itoa())
}

func StartServer() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./internal/app"))
	mux.Handle("/app/", http.StripPrefix("/app", fs))
	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app", fs)))
	mux.HandleFunc("/healthz", serveStatus)
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
