package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go1f/pkg/auth"
)

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, SignInResponse{Error: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, SignInResponse{Error: "Invalid JSON format"}, http.StatusBadRequest)
		return
	}

	envPassword := os.Getenv("TODO_PASSWORD")
	if envPassword == "" {
		log.Println("WARNING: TODO_PASSWORD not set, using default")
		envPassword = "defaultpassword"
	}

	log.Printf("Received password: '%s', expected: '%s'", req.Password, envPassword)

	if req.Password != envPassword {
		writeJSON(w, SignInResponse{Error: "Invalid password"}, http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		log.Printf("Token generation error: %v", err)
		writeJSON(w, SignInResponse{Error: "Internal server error"}, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(8 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeJSON(w, SignInResponse{Token: token}, http.StatusOK)
}
