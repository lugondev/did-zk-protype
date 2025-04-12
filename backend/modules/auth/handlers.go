package auth

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"did-example/modules/db"
	"did-example/modules/did"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrDIDNotFound        = errors.New("DID not found")
	ErrDIDVerification    = errors.New("DID verification failed")
)

// Global DID service instance
var didService *did.DIDService

func generateToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func verifyToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, ErrInvalidToken
}

// Initialize DID service
func InitDIDService(service *did.DIDService) {
	didService = service
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var user User
	err = db.GetDB().QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, created_at",
		req.Username,
		string(hashedPassword),
	).Scan(&user.ID, &user.Username, &user.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, ErrUserExists.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log login attempt
	fmt.Printf("Login attempt for username: %s\n", req.Username)

	var user User
	var hashedPassword string
	var emailNull, didNull sql.NullString
	err := db.GetDB().QueryRow(
		"SELECT id, username, password_hash, created_at, email, did, last_login_at, last_updated_at, is_two_factor_enabled FROM users WHERE username = $1",
		req.Username,
	).Scan(
		&user.ID,
		&user.Username,
		&hashedPassword,
		&user.CreatedAt,
		&emailNull,
		&didNull,
		&user.LastLoginAt,
		&user.LastUpdatedAt,
		&user.IsTwoFactorEnabled,
	)

	// Convert nullable strings to empty strings if null
	if emailNull.Valid {
		user.Email = emailNull.String
	}
	if didNull.Valid {
		user.DID = didNull.String
	}

	if err == sql.ErrNoRows {
		fmt.Printf("User not found: %s\n", req.Username)
		http.Error(w, ErrInvalidCredentials.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Printf("Database error during login: %v\n", err)
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		fmt.Printf("Invalid password for user: %s\n", req.Username)
		http.Error(w, ErrInvalidCredentials.Error(), http.StatusUnauthorized)
		return
	}

	// Update last login time
	_, err = db.GetDB().Exec(
		"UPDATE users SET last_login_at = NOW() WHERE id = $1",
		user.ID,
	)
	if err != nil {
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	token, err := generateToken(user.ID)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		http.Error(w, fmt.Sprintf("Token generation error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LoginWithDID handles authentication using DID
func LoginWithDID(w http.ResponseWriter, r *http.Request) {
	if didService == nil {
		http.Error(w, "DID service not initialized", http.StatusInternalServerError)
		return
	}

	var req DIDLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify the DID authentication
	signature, err := hex.DecodeString(req.Signature)
	if err != nil {
		http.Error(w, "Invalid signature format", http.StatusBadRequest)
		return
	}

	proof, err := hex.DecodeString(req.Proof)
	if err != nil {
		http.Error(w, "Invalid proof format", http.StatusBadRequest)
		return
	}

	valid, err := didService.VerifyAuthentication(req.DIDID, proof, signature)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, ErrDIDVerification.Error(), http.StatusUnauthorized)
		return
	}

	// Find user by DID
	var user User
	var emailNull sql.NullString
	err = db.GetDB().QueryRow(
		"SELECT id, username, email, did, created_at, last_login_at, last_updated_at, is_two_factor_enabled FROM users WHERE did = $1",
		req.DIDID,
	).Scan(
		&user.ID,
		&user.Username,
		&emailNull,
		&user.DID,
		&user.CreatedAt,
		&user.LastLoginAt,
		&user.LastUpdatedAt,
		&user.IsTwoFactorEnabled,
	)

	if emailNull.Valid {
		user.Email = emailNull.String
	}

	if err == sql.ErrNoRows {
		http.Error(w, ErrDIDNotFound.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update last login time
	_, err = db.GetDB().Exec(
		"UPDATE users SET last_login_at = NOW() WHERE id = $1",
		user.ID,
	)
	if err != nil {
		// Log the error but don't fail the login
		println("Failed to update last login time:", err)
	}

	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var user User
	var emailNull, didNull sql.NullString
	err := db.GetDB().QueryRow(`
    SELECT id, username, email, did, created_at, last_login_at, last_updated_at, is_two_factor_enabled 
    FROM users WHERE id = $1
  `, userID).Scan(
		&user.ID,
		&user.Username,
		&emailNull,
		&didNull,
		&user.CreatedAt,
		&user.LastLoginAt,
		&user.LastUpdatedAt,
		&user.IsTwoFactorEnabled,
	)

	// Convert nullable strings to empty strings if null
	if emailNull.Valid {
		user.Email = emailNull.String
	}
	if didNull.Valid {
		user.DID = didNull.String
	}

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build update query based on provided fields
	query := "UPDATE users SET last_updated_at = NOW()"
	params := []interface{}{}
	paramCount := 1

	if req.Username != nil {
		// Verify username uniqueness
		var exists bool
		err := db.GetDB().QueryRow(
			"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)",
			*req.Username, userID,
		).Scan(&exists)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}

		query += fmt.Sprintf(", username = $%d", paramCount)
		params = append(params, *req.Username)
		paramCount++
	}

	if req.Email != nil {
		query += fmt.Sprintf(", email = $%d", paramCount)
		params = append(params, *req.Email)
		paramCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", paramCount)
	params = append(params, userID)

	// Execute update
	result, err := db.GetDB().Exec(query, params...)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get update result", http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return updated user profile
	GetUserProfile(w, r)
}

func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify current password
	var hashedPassword string
	err := db.GetDB().QueryRow(
		"SELECT password_hash FROM users WHERE id = $1",
		userID,
	).Scan(&hashedPassword)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash and store new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = db.GetDB().Exec(
		"UPDATE users SET password_hash = $1, last_updated_at = NOW() WHERE id = $2",
		string(newHashedPassword), userID,
	)

	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	response := NewSuccessResponse(nil, "Password updated successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateSecurity(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req UpdateSecurityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update security settings
	_, err := db.GetDB().Exec(
		"UPDATE users SET is_two_factor_enabled = $1, last_updated_at = NOW() WHERE id = $2",
		req.EnableTwoFactor, userID,
	)

	if err != nil {
		http.Error(w, "Failed to update security settings", http.StatusInternalServerError)
		return
	}

	status := map[bool]string{true: "enabled", false: "disabled"}[req.EnableTwoFactor]
	response := NewSuccessResponse(nil, fmt.Sprintf("Two-factor authentication is now %s", status))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		userID, err := verifyToken(bearerToken[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", fmt.Sprint(userID))
		next.ServeHTTP(w, r)
	}
}

func InitDB() error {
	// Add migration for new user fields if needed
	_, err := db.GetDB().Exec(`
    ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS email TEXT,
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS last_updated_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS is_two_factor_enabled BOOLEAN DEFAULT false;
  `)
	return err
}
