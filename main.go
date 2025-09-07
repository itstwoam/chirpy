package main

import (
	"fmt"
	"github.com/itstwoam/internal/server"
)

func main() {
	server.StartServer()
	fmt.Println("I think it started.")
}
