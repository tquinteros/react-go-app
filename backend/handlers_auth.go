package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	var user User
	err = conn.QueryRow(context.Background(),
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email",
		req.Email, string(hash),
	).Scan(&user.ID, &user.Email)
	if err != nil {
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}
	accessToken, err := generateAccessToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	refreshToken, _ := generateRefreshToken(user.ID, user.Email)
	authCookie(w, refreshToken, 7*24*60*60)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{AccessToken: accessToken, User: user})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	var user User
	var passwordHash string
	err := conn.QueryRow(context.Background(),
		"SELECT id, email, password_hash FROM users WHERE email = $1", req.Email,
	).Scan(&user.ID, &user.Email, &passwordHash)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	accessToken, err := generateAccessToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	refreshToken, _ := generateRefreshToken(user.ID, user.Email)
	authCookie(w, refreshToken, 7*24*60*60)
	json.NewEncoder(w).Encode(AuthResponse{AccessToken: accessToken, User: user})
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "no refresh token", http.StatusUnauthorized)
		return
	}
	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	userIDVal, ok := claims["user_id"]
	if !ok || userIDVal == nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}
	userID := int(userIDVal.(float64))
	var email string
	if emailVal, ok := claims["email"]; ok && emailVal != nil {
		email = emailVal.(string)
	} else {
		if err := conn.QueryRow(context.Background(),
			"SELECT email FROM users WHERE id = $1", userID).Scan(&email); err != nil {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
			return
		}
	}
	accessToken, err := generateAccessToken(userID, email)
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authCookie(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}
