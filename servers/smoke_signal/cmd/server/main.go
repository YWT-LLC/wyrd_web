/* wyrd_web
 * Copyright (c) 2025 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
  "github.com/joho/godotenv"

	"smoke-signal/internal/db"
	"smoke-signal/internal/handlers"
)

func main() {
	godotenv.Load()

	if err := db.Connect(); err != nil {
		log.Fatal("DB connect failed:", err)
	}

	r := mux.NewRouter()

	// Public routes
	http.HandleFunc("/signUp", handlers.SignUp)
	http.HandleFunc("/login", handlers.Login)

	// Protected API
	api := r.PathPrefix("/api").Subrouter()
	api.Use(
		middleware.RecoveryMiddleware,
		middleware.CORSMiddleware,
		middleware.LoggingMiddleware,
		middleware.AuthMiddleware,
	)

	api.HandleFunc("/me", handlers.Me).Methods("GET")
	api.HandleFunc("/groups", handlers.CreateGroup).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
