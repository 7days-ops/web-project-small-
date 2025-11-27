package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var jwtSecret []byte

// Rate limiting: максимум 5 попыток в 15 минут с одного IP
var rateLimitMap = make(map[string][]time.Time)
var rateLimitMutex sync.Mutex

// Валидация пароля: минимум 8 символов
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	return nil
}

// Валидация имени пользователя
func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 32 {
		return fmt.Errorf("username must be no more than 32 characters")
	}
	return nil
}

// Rate limiting: проверка количества попыток с одного IP
func checkRateLimit(ip string) error {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()

	now := time.Now()
	fifteenMinutesAgo := now.Add(-15 * time.Minute)

	// Очищаем старые попытки
	if attempts, exists := rateLimitMap[ip]; exists {
		var recentAttempts []time.Time
		for _, attempt := range attempts {
			if attempt.After(fifteenMinutesAgo) {
				recentAttempts = append(recentAttempts, attempt)
			}
		}
		rateLimitMap[ip] = recentAttempts

		if len(recentAttempts) >= 5 {
			return fmt.Errorf("too many attempts, please try again later")
		}
	}

	// Добавляем новую попытку
	rateLimitMap[ip] = append(rateLimitMap[ip], now)
	return nil
}

// Получение IP адреса клиента
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	return r.RemoteAddr
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "/data/auth.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
	// Только localhost:3000 разрешен
	allowedOrigin := "http://localhost:3000"
	originHeader := r.Header.Get("Origin")

	if originHeader == allowedOrigin {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	clientIP := getClientIP(r)
	log.Printf("[REGISTER] %s %s from %s", r.Method, r.URL.Path, clientIP)

	// Проверка rate limiting
	if err := checkRateLimit(clientIP); err != nil {
		log.Printf("[REGISTER] Rate limit exceeded for %s: %v", clientIP, err)
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[REGISTER] Decode error: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Валидация входных данных
	if err := validateUsername(req.Username); err != nil {
		log.Printf("[REGISTER] Invalid username: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validatePassword(req.Password); err != nil {
		log.Printf("[REGISTER] Invalid password: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[REGISTER] Hash error: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", req.Username, string(hashedPassword))
	if err != nil {
		log.Printf("[REGISTER] Insert error: %v", err)
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	id, _ := result.LastInsertId()
	user := User{ID: int(id), Username: req.Username}

	token, err := generateToken(user)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{Token: token, User: user}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func login(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	clientIP := getClientIP(r)
	log.Printf("[LOGIN] %s %s from %s", r.Method, r.URL.Path, clientIP)

	// Проверка rate limiting
	if err := checkRateLimit(clientIP); err != nil {
		log.Printf("[LOGIN] Rate limit exceeded for %s: %v", clientIP, err)
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user User
	var hashedPassword string
	err := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &hashedPassword)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{Token: token, User: user}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func verify(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}

	tokenString := authHeader[7:] // Remove "Bearer "

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":    true,
		"user_id":  claims.UserID,
		"username": claims.Username,
	})
}

func generateToken(user User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func main() {
	// Инициализация JWT секрета из переменной окружения
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		log.Println("WARNING: JWT_SECRET not set in environment, using default (NOT SECURE FOR PRODUCTION)")
		secretKey = "super-secret-key-change-in-production"
	}
	jwtSecret = []byte(secretKey)

	initDB()
	defer db.Close()

	log.Println("Setting up routes...")
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ROUTE] Received %s %s", r.Method, r.URL.Path)
		register(w, r)
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ROUTE] Received %s %s", r.Method, r.URL.Path)
		login(w, r)
	})
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ROUTE] Received %s %s", r.Method, r.URL.Path)
		verify(w, r)
	})

	fmt.Println("Auth Service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
