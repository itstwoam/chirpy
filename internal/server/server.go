package server

import (
	"net/http"
	"fmt"
)

func StartServer() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./internal/app"))
	mux.Handle("/app/", http.StripPrefix("/app", fs))
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
