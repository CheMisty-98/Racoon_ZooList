package handlers

import (
	"Zoo_List/internal/config"
	_ "Zoo_List/internal/middleware"
	"Zoo_List/internal/models"
	"Zoo_List/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type Handler struct {
	DB *sql.DB
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ZooList is working!")
}

type RegisterRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var req RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteInvalidJSON(w)
			return
		}

		fmt.Printf("Registering user: %s\n", req.Nickname)

		var exists bool
		err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE nickname = ?)", req.Nickname).Scan(&exists)
		if err != nil {
			utils.WriteError400(w, "Invalid data")
			return
		}

		if exists {
			utils.WriteAlreadyExists(w, "User")
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.WriteError500(w, "Error processing password")
			return
		}

		_, err = h.DB.Exec("INSERT INTO users(nickname, password_hash) VALUES (?, ?)", req.Nickname, string(passwordHash))
		if err != nil {
			utils.WriteError500(w, "Error saving user")
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "User registered successfully")
	}
}

type LoginRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var req LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteInvalidJSON(w)
			return
		}

		fmt.Printf("Logining user: %s\n", req.Nickname)

		var dbPasswordHash string
		err = h.DB.QueryRow("SELECT password_hash FROM users WHERE nickname = ?", req.Nickname).Scan(&dbPasswordHash)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.WriteUserNotFound(w)
			} else {
				utils.WriteDBError(w)
			}
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(dbPasswordHash), []byte(req.Password))
		if err != nil {
			utils.WriteInvalidPassword(w)
			return
		}

		var userID int
		err = h.DB.QueryRow("SELECT id FROM users WHERE nickname = ?", req.Nickname).Scan(&userID)

		cfg := config.Load()
		token, err := generateToken(userID, cfg.JWTSecret)
		if err != nil {
			utils.WriteError500(w, "Error generating token")
			return
		}

		response := map[string]string{
			"message": "Login successful",
			"token":   token,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		return
	}
}

func generateToken(userID int, secret string) (string, error) {
	claims := &models.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
