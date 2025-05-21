package main

import (
	"log"
	"net/http"
	"os"

	"smoke-signal/internal/db"
	"smoke-signal/internal/handlers"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	if err := db.Connect(); err != nil {
		log.Fatal("DB connect failed:", err)
	}

	http.HandleFunc("/signup", handlers.SignUp)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
