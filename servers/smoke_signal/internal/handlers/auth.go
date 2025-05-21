/* wyrd_web
 * Copyright (c) 2025 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */
 
package handlers

import (
	"net/http"
	"smoke-signal/internal/auth"
	"smoke-signal/internal/db"
	"smoke-signal/internal/models"

	"encoding/json"
)

type CredsInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var input CredsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashed, err := auth.HashPassword(input.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: hashed,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateJWT(user.ID.String())
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input CredsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPasswordHash(input.Password, user.Password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(user.ID.String())
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Invalidate the JWT token here if needed
	// For stateless JWT, you might not need to do anything
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
