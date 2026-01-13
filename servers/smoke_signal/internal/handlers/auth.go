/* wyrd_web
 * Copyright (c) 2026 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */
 
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
	
	"gorm.io/gorm"

	"smoke-signal/internal/auth"
	"smoke-signal/internal/db"
	"smoke-signal/internal/models"
)

type CredsInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	ID    string `json:"id"`
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
		Email:    strings.ToLower(input.Email),
		Password: hashed,
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		profile.UserID = user.ID
		return tx.Create(&profile).Error
	})

	if err != nil {
		http.Error(w, "User creation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	token, err := auth.GenerateJWT(user.ID.String())
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	var resp AuthResponse
	resp.Token = token
	resp.User.ID = user.ID.String()

  json.NewEncoder(w).Encode(resp)
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

	var resp AuthResponse
	resp.Token = token
	resp.User.ID = user.ID.String()

  json.NewEncoder(w).Encode(resp)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Invalidate the JWT token here if needed
	// For stateless JWT, you might not need to do anything
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
