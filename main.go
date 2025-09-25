package main

import (
	"fmt"
	"github.com/itstwoam/chirpy/internal/server"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	//"github.com/google/uuid"
	"os"
	//"database/sql"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	key := os.Getenv("KEY")
	polka_key := os.Getenv("POLKA_KEY")
	//db, _ := sql.Open("postgres", dbURL)
	//db, err := sql.Open("postgres", dbURL)
	//_ := database.New(db)
	//dbQueries := database.New(db)
	server.StartServer(dbURL, platform, key, polka_key)
	fmt.Println("I think it started.")
}
