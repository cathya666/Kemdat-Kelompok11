package handler

import (
	"aaa_bekend/internal/auth"
	"aaa_bekend/internal/db"
	"encoding/json"
	"net/http"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	ok, err := auth.CheckPassword(req.Password, user.PasswordHash)
	if err != nil || !ok {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateToken(user)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func Register(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "student"
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = db.CreateUser(r.Context(), req.Username, hash, req.Role)
	if err != nil {
		http.Error(w, "User already exists or DB Error", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{"message": "User is created"})
}
