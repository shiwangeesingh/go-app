package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"os"
	"io/ioutil"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"fmt"
	"github.com/shiwangeesingh/go-app/internal/db"
	// "github.com/shiwangeesingh/go-app/utils"
	"github.com/shiwangeesingh/go-app/users"
	"golang.org/x/crypto/bcrypt"
	"github.com/shiwangee/go-app/middleware"

)

type user struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int32  `json:"age"`
	Gender string `json:"gender"`
	Password  []byte `json:"password"` 
}

var queries *db.Queries
var conn *sql.DB // Store DB connection separately

// func main() {
// 	// Connect to Postgres
// 	var err error

// 	// dbHost := os.Getenv("DB_HOST")
// 	// dbUser := os.Getenv("DB_USER")
// 	// dbPassword := os.Getenv("DB_PASSWORD")
// 	// dbName := os.Getenv("DB_NAME")
// 	// dbPort := os.Getenv("DB_PORT")

// 	// // Create connection string
// 	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
// 	// 	dbHost, dbUser, dbPassword, dbName, dbPort)

// 	// // Connect to the database
// 	// conn, err := sql.Open("postgres", dsn)
// 	//  conn, err = sql.Open("postgres", "postgres://admin:admin@localhost:5432/users_survey?sslmode=disable")
// 	 conn, err = sql.Open("postgres", "postgres://admin:admin@db:5432/users_survey?sslmode=disable")

// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	defer conn.Close()

// 	// Ensure DB is ready before executing schema
// 	err = conn.Ping()
// 	if err != nil {
// 		log.Fatalf("Database not ready: %v", err)
// 	}

// 	// Read and execute schema.sql
// 	schemaFile := "schema.sql"
// 	schema, err := ioutil.ReadFile(schemaFile)
// 	if err != nil {
// 		log.Fatalf("Failed to read schema file: %v", err)
// 	}

// 	_, err = conn.Exec(string(schema))
// 	if err != nil {
// 		log.Fatalf("Failed to execute schema: %v", err)
// 	}

// 	log.Println("Database schema applied successfully!")

// 	// Initialize sqlc queries
// 	// queries = db.New(db)
// 	queries = db.New(conn)

// 	// Setup Router
// 	r := chi.NewRouter()
// 	r.Get("/users", getUsers)
// 	r.Post("/users", createUser)
// 	r.Delete("/users/{id}", deleteUser)
// 	r.Put("/users/{id}", updateUser)

// 	r.Mount("/users-api", users.Routes()) // Mount user-related routes

// 	// Start Server
// 	log.Println("0.0.0.0:8080")
// 	// http.ListenAndServe("localhost:8080", r)
// 	http.ListenAndServe("0.0.0.0:8080", r)

// }

// Global database connection

func main() {
	var err error

	// Use environment variables or default values
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "admin")
	dbPassword := getEnv("DB_PASSWORD", "admin")
	dbName := getEnv("DB_NAME", "users_survey")
	dbPort := getEnv("DB_PORT", "5432")

	// Create DSN connection string
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Connect to the database
	conn, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Ensure DB is ready
	if err = conn.Ping(); err != nil {
		log.Fatalf("❌ Database not ready: %v", err)
	}

	log.Println("✅ Connected to database!")

	// Apply schema
	if err = applySchema("schema.sql"); err != nil {
		log.Fatalf("❌ Failed to execute schema: %v", err)
	}

	// Initialize sqlc queries
	queries = db.New(conn)

	// Setup router
	r := chi.NewRouter()

	// Mount user-related routes
	r.Mount("/users-api", users.Routes())
	// r.Get("/users", getUsers)
	r.Post("/users", createUser)
	r.Post("/login", login)
	// r.Delete("/users/{id}", deleteUser)
	// r.Put("/users/{id}", updateUser)
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.AuthMiddleware)
		protected.Get("/users", getUsers)
		protected.Delete("/users/{id}", deleteUser)
		protected.Put("/users/{id}", updateUser)
	})
	// Start Server
	log.Println("🚀 Server running on 0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// getEnv gets an environment variable or fallback to a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// applySchema reads and applies schema.sql
func applySchema(schemaFile string) error {
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}
	_, err = conn.Exec(string(schema))
	return err
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
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
		Password: hashedPassword,
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


func login(w http.ResponseWriter, r *http.Request) {
	//	queries := db.New(db.DB) // Initialize SQLC queries
	var newUser user
	
		// Decode JSON request
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		user, err := queries.GetUserByEmail(r.Context(), newUser.Email)
		if err != nil {
			 http.Error(w, "user not found", http.StatusBadRequest)
			 return;
		}
	
		// Compare hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newUser.Password)); err != nil {
			 http.Error(w, "invalid password", http.StatusBadRequest)
			 return;
	
		} 
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
