package main

import (
	"fmt"
	"github.com/itstwoam/chirpy/internal/server"
)

func main() {
	server.StartServer()
	fmt.Println("I think it started.")
}
