package users

import (
	"encoding/json"
	"net/http"

	"github.com/shiwangeesingh/goFirstProject/internal/db"
	"github.com/shiwangeesingh/goFirstProject/config/utils"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Name string `json:"name"`
	Password string `json:"password"`
}

type user struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int32  `json:"age"`
	Gender string `json:"gender"`
	Password string `json:"password"`
}

var queries *db.Queries

// Register a new user
// func RegisterHandler(w http.ResponseWriter, r *http.Request) {
// 	var creds Credentials
// 	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		http.Error(w, "Error hashing password", http.StatusInternalServerError)
// 		return
// 	}

// 	//////////////////////////_, err = db.DB.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", creds.Username, string(hashedPassword))
// 	if err != nil {
// 		http.Error(w, "Error creating user", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// }

func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser user
	log.Printf("Register user is called")

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
		Password: hashedPassword
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

// Login user and return JWT token
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var storedHash string
	err := db.DB.Get(&storedHash, "SELECT password_hash FROM users WHERE username=$1", creds.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(creds.Name)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}



func AuthenticateUser(ctx context.Context, email, password string) (*db.User, error) {
	queries := db.New(db.DB) // Initialize SQLC queries

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
