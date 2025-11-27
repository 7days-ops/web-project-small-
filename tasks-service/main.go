package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Task struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

func initDB() {
	var err error
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = ""
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "task_manager"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		status VARCHAR(50) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func verifyToken(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("no token provided")
	}

	// Create request to auth service
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8080"
	}
	req, err := http.NewRequest("GET", authServiceURL+"/verify", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("invalid token")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	userID := int(result["user_id"].(float64))
	return userID, nil
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	userID, err := verifyToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.Query("SELECT id, user_id, title, description, status, created_at, updated_at FROM tasks WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	userID, err := verifyToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO tasks (user_id, title, description, status) VALUES (?, ?, ?, ?)",
		userID, req.Title, req.Description, "pending")
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	task := Task{
		ID:          int(id),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	userID, err := verifyToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/tasks/"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE tasks SET title = COALESCE(NULLIF(?, ''), title), description = COALESCE(NULLIF(?, ''), description), status = COALESCE(NULLIF(?, ''), status), updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
		req.Title, req.Description, req.Status, taskID, userID)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var task Task
	err = db.QueryRow("SELECT id, user_id, title, description, status, created_at, updated_at FROM tasks WHERE id = ?", taskID).
		Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	userID, err := verifyToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/tasks/"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("DELETE FROM tasks WHERE id = ? AND user_id = ?", taskID, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tasks" {
		switch r.Method {
		case "GET":
			getTasks(w, r)
		case "POST":
			createTask(w, r)
		case "OPTIONS":
			enableCORS(w)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(r.URL.Path, "/tasks/") {
		switch r.Method {
		case "PUT":
			updateTask(w, r)
		case "DELETE":
			deleteTask(w, r)
		case "OPTIONS":
			enableCORS(w)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/", tasksHandler)

	fmt.Println("Tasks Service running on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
