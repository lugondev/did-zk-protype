// Package main implements a DID (Decentralized Identifier) HTTP server
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"did-example/modules/did"
)

// CreateDIDRequest represents the request body for DID creation
type CreateDIDRequest struct {
	Name string `json:"name"`
	DOB  string `json:"dob"` // Date of birth as string to match frontend type
}

// CreateDIDResponse represents the response body after DID creation
type CreateDIDResponse struct {
	DID        *did.DID `json:"did"`        // The created DID object
	PrivateKey string   `json:"privateKey"` // Private key in hex format
}

// AuthRequest represents the request body for DID authentication
type AuthRequest struct {
	DIDID      string `json:"didId"`      // DID identifier
	PrivateKey string `json:"privateKey"` // Private key in hex format
	Challenge  string `json:"challenge"`  // Authentication challenge
}

// AuthResponse represents the response body after DID authentication
type AuthResponse struct {
	Proof     string `json:"proof"`     // ZKP proof
	Signature string `json:"signature"` // EdDSA signature
}

// VerifyRequest represents the request body for verifying DID authentication
type VerifyRequest struct {
	DIDID     string `json:"didId"`     // DID identifier
	Signature string `json:"signature"` // EdDSA signature to verify
	Proof     string `json:"proof"`     // ZKP proof to verify
}

// VerifyResponse represents the response body after verification
type VerifyResponse struct {
	Verified bool `json:"verified"` // Whether the verification was successful
}

// main initializes and starts the HTTP server with DID functionality
func main() {
	// Initialize DID service for handling decentralized identity operations
	didService, err := did.NewDIDService()
	if err != nil {
		log.Fatalf("Failed to initialize DID service: %v", err)
	}

	// Handler for creating a new DID with age credentials
	http.HandleFunc("/api/did/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateDIDRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Convert date of birth string to integer
		dob, err := strconv.Atoi(req.DOB)
		if err != nil {
			http.Error(w, "Invalid DOB format", http.StatusBadRequest)
			return
		}

		// Create new DID with EdDSA key pair
		newDID, privateKey, err := didService.CreateDID()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create DID: %v", err), http.StatusInternalServerError)
			return
		}

		// Issue age credential using zero-knowledge proofs
		if _, _, err := didService.IssueAgeCredential(newDID.ID, dob); err != nil {
			http.Error(w, fmt.Sprintf("Failed to issue age credential: %v", err), http.StatusInternalServerError)
			return
		}

		// Format private key for response
		seed := make([]byte, 32)
		privKeyBytes := privateKey.Bytes()
		copy(seed[32-len(privKeyBytes):], privKeyBytes)

		response := CreateDIDResponse{
			DID:        newDID,
			PrivateKey: hex.EncodeToString(seed),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Handler for authenticating a DID using zero-knowledge proofs
	http.HandleFunc("/api/did/authenticate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Convert hex-encoded private key to big.Int
		privateKey := new(big.Int)
		_, success := privateKey.SetString(req.PrivateKey, 16)
		if !success {
			http.Error(w, "Invalid private key format", http.StatusBadRequest)
			return
		}

		// Generate ZKP proof and EdDSA signature
		proof, signature, err := didService.AuthenticateDID(req.DIDID, privateKey, req.Challenge)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := AuthResponse{
			Proof:     hex.EncodeToString(proof),
			Signature: hex.EncodeToString(signature),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Handler for verifying DID authentication proofs
	http.HandleFunc("/api/did/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req VerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Decode hex-encoded proof and signature
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

		// Verify both ZKP proof and EdDSA signature
		valid, err := didService.VerifyAuthentication(req.DIDID, proof, signature)
		if err != nil {
			http.Error(w, fmt.Sprintf("Verification failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := VerifyResponse{
			Verified: valid,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Apply CORS middleware for cross-origin requests
	handler := corsMiddleware(http.DefaultServeMux)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// corsMiddleware handles Cross-Origin Resource Sharing (CORS) headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
