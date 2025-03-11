package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
//	"os"
	"io/ioutil"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
//	"fmt"
	"github.com/shiwangeesingh/goFirstProject/internal/db"
	"github.com/shiwangeesingh/go-app/config/utils"
	"github.com/shiwangeesingh/go-app/config/users"
)

type user struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int32  `json:"age"`
	Gender string `json:"gender"`
}

var queries *db.Queries
var conn *sql.DB // Store DB connection separately

func main() {
	// Connect to Postgres
	var err error

	// dbHost := os.Getenv("DB_HOST")
	// dbUser := os.Getenv("DB_USER")
	// dbPassword := os.Getenv("DB_PASSWORD")
	// dbName := os.Getenv("DB_NAME")
	// dbPort := os.Getenv("DB_PORT")

	// // Create connection string
	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
	// 	dbHost, dbUser, dbPassword, dbName, dbPort)

	// // Connect to the database
	// conn, err := sql.Open("postgres", dsn)
	//  conn, err = sql.Open("postgres", "postgres://admin:admin@localhost:5432/users_survey?sslmode=disable")
	 conn, err = sql.Open("postgres", "postgres://admin:admin@db:5432/users_survey?sslmode=disable")

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Ensure DB is ready before executing schema
	err = conn.Ping()
	if err != nil {
		log.Fatalf("Database not ready: %v", err)
	}

	// Read and execute schema.sql
	schemaFile := "schema.sql"
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	_, err = conn.Exec(string(schema))
	if err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	log.Println("Database schema applied successfully!")

	// Initialize sqlc queries
	// queries = db.New(db)
	queries = db.New(conn)

	// Setup Router
	r := chi.NewRouter()
	r.Get("/users", getUsers)
	r.Post("/users", createUser)
	r.Delete("/users/{id}", deleteUser)
	r.Put("/users/{id}", updateUser)
	r.Post("/registerUser", user.createUser)


	// Start Server
	log.Println("0.0.0.0:8080")
	// http.ListenAndServe("localhost:8080", r)
	http.ListenAndServe("0.0.0.0:8080", r)

}

// 🔹 Corrected `getGeneration` function
func getGeneration(age int32) (string, string) {
	switch {
	case age <= 12:
		return "Generation Alpha", "C"
	case age <= 28:
		return "Generation Z", "B"
	case age <= 44:
		return "Millennials", "A"
	default:
		return "Unknown", ""
	}
}

// 🔹 Get All Users
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	users, err := queries.GetUsers(r.Context())
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

// 🔹 Create User with Transaction Handling
func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser user

	// Decode JSON request
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Received User: %+v\n", newUser)
	if err := conn.Ping(); err != nil {
		log.Fatal("Database is not reachable:", err)
	}
	
	// Start transaction
	tx, err := conn.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}


	// Use transaction
	qtx := queries.WithTx(tx)

	// Insert User
	userID, err := qtx.InsertUser(r.Context(), db.InsertUserParams{
		Name:   newUser.Name,
		Age:    newUser.Age,
		Gender: newUser.Gender,
		Email:  newUser.Email,
	})
	log.Printf("userID", userID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		return
	}

	// Determine Generation
	userGeneration, grade := getGeneration(newUser.Age)
	log.Printf("userGeneration, grade", userID, userGeneration, grade)
	if userGeneration != "" {
		// gradeNull := sql.NullString{String: grade, Valid: grade != ""} // Convert to sql.NullString
		log.Printf("userGeneration: %s, grade: %s", userGeneration, grade)
		log.Printf("userID before InsertGeneration: %d", userID)
		_, err = qtx.InsertGeneration(r.Context(), db.InsertGenerationParams{
			UserID:     userID,
			Generation: userGeneration, // Ensure this is a valid enum
			Grade:      grade,                          // Consider sql.NullString{String: grade, Valid: grade != ""}
		})
		if err != nil {
			log.Printf("InsertGeneration error: %v", err) // Log the error
			tx.Rollback()                                 // Rollback on error
			http.Error(w, "Failed to insert user generation", http.StatusInternalServerError)
			return
		}

	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	newUser.ID = int(userID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

// 🔹 Delete User
func deleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = queries.DeleteUser(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

// 🔹 Update User
func updateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updatedUser user
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = queries.UpdateUser(r.Context(), db.UpdateUserParams{
		ID:     int32(id),
		Name:   updatedUser.Name,
		Age:    updatedUser.Age,
		Gender: updatedUser.Gender,
		Email:  updatedUser.Email,
	})
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedUser)
}
