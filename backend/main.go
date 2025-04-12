package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"did-example/modules/auth"
	"did-example/modules/db"
	"did-example/modules/did"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize database connection and schema
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize DID service
	didService, err := did.NewDIDService()
	if err != nil {
		log.Fatalf("Failed to initialize DID service: %v", err)
	}

	// Initialize auth service with DID service
	auth.InitDIDService(didService)

	// Auth routes
	http.HandleFunc("/api/auth/register", auth.Register)
	http.HandleFunc("/api/auth/login", auth.Login)
	http.HandleFunc("/api/auth/login-with-did", auth.LoginWithDID)

	// User profile and settings routes
	http.HandleFunc("/api/users/me", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth.GetUserProfile(w, r)
		case http.MethodPut, http.MethodPatch:
			auth.UpdateProfile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/users/me/password", auth.AuthMiddleware(auth.UpdatePassword))
	http.HandleFunc("/api/users/me/security", auth.AuthMiddleware(auth.UpdateSecurity))

	// Protected DID routes with auth middleware
	http.HandleFunc("/api/did/create", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
			DOB  string `json:"dob"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dob, err := strconv.Atoi(req.DOB)
		if err != nil {
			http.Error(w, "Invalid DOB format", http.StatusBadRequest)
			return
		}

		newDID, privateKey, err := didService.CreateDID()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create DID: %v", err), http.StatusInternalServerError)
			return
		}

		if _, _, err := didService.IssueAgeCredential(newDID.ID, dob); err != nil {
			http.Error(w, fmt.Sprintf("Failed to issue age credential: %v", err), http.StatusInternalServerError)
			return
		}

		// Update user's DID in database
		userID := r.Header.Get("X-User-ID")
		if _, err := db.GetDB().Exec("UPDATE users SET did = $1 WHERE id = $2", newDID.ID, userID); err != nil {
			http.Error(w, "Failed to update user DID", http.StatusInternalServerError)
			return
		}

		seed := make([]byte, 32)
		privKeyBytes := privateKey.Bytes()
		copy(seed[32-len(privKeyBytes):], privKeyBytes)

		response := struct {
			DID        *did.DID `json:"did"`
			PrivateKey string   `json:"privateKey"`
		}{
			DID:        newDID,
			PrivateKey: hex.EncodeToString(seed),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	http.HandleFunc("/api/did/authenticate", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			DIDID      string `json:"didId"`
			PrivateKey string `json:"privateKey"`
			Challenge  string `json:"challenge"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		privateKey := new(big.Int)
		_, success := privateKey.SetString(req.PrivateKey, 16)
		if !success {
			http.Error(w, "Invalid private key format", http.StatusBadRequest)
			return
		}

		proof, signature, err := didService.AuthenticateDID(req.DIDID, privateKey, req.Challenge)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := struct {
			Proof     string `json:"proof"`
			Signature string `json:"signature"`
		}{
			Proof:     hex.EncodeToString(proof),
			Signature: hex.EncodeToString(signature),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	http.HandleFunc("/api/did/verify", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			DIDID     string `json:"didId"`
			Signature string `json:"signature"`
			Proof     string `json:"proof"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		proof, err := hex.DecodeString(req.Proof)
		if err != nil {
			http.Error(w, "Invalid proof format", http.StatusBadRequest)
			return
		}

		signature, err := hex.DecodeString(req.Signature)
		if err != nil {
			http.Error(w, "Invalid signature format", http.StatusBadRequest)
			return
		}

		valid, err := didService.VerifyAuthentication(req.DIDID, proof, signature)
		if err != nil {
			http.Error(w, fmt.Sprintf("Verification failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := struct {
			Verified bool `json:"verified"`
		}{
			Verified: valid,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Apply CORS middleware
	handler := corsMiddleware(http.DefaultServeMux)

	// Print server initialization message
	log.Println("Server initializing...")
	log.Println("Creating database schema...")
	log.Println("Setting up routes...")
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
