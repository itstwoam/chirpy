package server

import (
	"net/http"
	"fmt"
)

func StartServer() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/", http.StripPrefix("/", fs))
	server := http.Server{}
	server.Handler = mux
	server.Addr = ":8085"
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("There was an error, but wtf?")
	}
	fmt.Println("Am I still running?")
}
