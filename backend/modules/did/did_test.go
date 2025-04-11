package did

import (
	"bytes"
	"testing"

	cryptomimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"

	"github.com/consensys/gnark-crypto/ecc"
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	cryptoeddsa "github.com/consensys/gnark-crypto/signature/eddsa"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

func TestDIDCircuit(t *testing.T) {
	// Test DID creation and authentication
	service, err := NewDIDService()
	if err != nil {
		t.Fatalf("Failed to create DID service: %v", err)
	}

	// Create a new DID
	did, privateKey, err := service.CreateDID()
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	if did.ID == "" {
		t.Error("DID ID should not be empty")
	}

	seed := make([]byte, 32)
	privKeyBytes := privateKey.Bytes()
	copy(seed[32-len(privKeyBytes):], privKeyBytes)

	privKey, err := cryptoeddsa.New(tedwards.BN254, bytes.NewReader(seed))
	if err != nil {
		t.Fatalf("Failed to create private key: %v", err)
	}

	// Test authentication
	challenge := "test authentication challenge"
	msg := []byte(challenge)
	hasher := cryptomimc.NewMiMC()

	signature, err := privKey.Sign(msg, hasher)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	proof, _, err := service.AuthenticateDID(did.ID, privateKey, challenge)
	if err != nil {
		t.Fatalf("Failed to authenticate DID: %v", err)
	}
	if proof == nil {
		t.Error("Proof should not be nil")
	}

	// Verify authentication
	valid, err := service.VerifyAuthentication(did.ID, proof, signature)
	if err != nil {
		t.Fatalf("Failed to verify authentication: %v", err)
	}
	if !valid {
		t.Error("Authentication verification should succeed")
	}

	// Test with wrong challenge
	wrongChallenge := []byte("wrong challenge")
	wrongSignature, err := privKey.Sign(wrongChallenge, hasher)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	valid, err = service.VerifyAuthentication(did.ID, proof, wrongSignature)
	if err != nil {
		t.Fatalf("Failed to verify authentication with wrong challenge: %v", err)
	}
	if valid {
		t.Error("Authentication verification should fail with wrong challenge")
	}
}

func TestAgeProofCircuit(t *testing.T) {
	// Test age credential issuance and verification
	service, err := NewDIDService()
	if err != nil {
		t.Fatalf("Failed to create DID service: %v", err)
	}

	// Create a DID
	did, _, err := service.CreateDID()
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	// Issue age credential
	actualAge := 25
	credential, salt, err := service.IssueAgeCredential(did.ID, actualAge)
	if err != nil {
		t.Fatalf("Failed to issue age credential: %v", err)
	}

	if credential.ID == "" {
		t.Error("Credential ID should not be empty")
	}

	// Create age proof
	ageThreshold := 18
	proof, err := service.CreateAgeProof(did.ID, credential.ID, ageThreshold, actualAge, salt)
	if err != nil {
		t.Fatalf("Failed to create age proof: %v", err)
	}

	// Verify age proof
	valid, err := service.VerifyAgeProof(did.ID, credential.ID, ageThreshold, proof)
	if err != nil {
		t.Fatalf("Failed to verify age proof: %v", err)
	}

	if !valid {
		t.Error("Age proof verification should succeed")
	}

	// Test with wrong age threshold
	wrongThreshold := 30
	valid, err = service.VerifyAgeProof(did.ID, credential.ID, wrongThreshold, proof)
	if err != nil {
		t.Fatalf("Failed to verify age proof with wrong threshold: %v", err)
	}

	if valid {
		t.Error("Age proof verification should fail with wrong threshold")
	}
}

func TestCircuitCompilation(t *testing.T) {
	// Test DIDAuthCircuit compilation
	var authCircuit DIDAuthCircuit
	cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &authCircuit)
	if err != nil {
		t.Fatalf("Failed to compile DIDAuthCircuit: %v", err)
	}

	_, _, err = groth16.Setup(cs)
	if err != nil {
		t.Fatalf("Failed to setup DIDAuthCircuit: %v", err)
	}

	// Test AgeProofCircuit compilation
	var ageProofCircuit AgeProofCircuit
	cs, err = frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &ageProofCircuit)
	if err != nil {
		t.Fatalf("Failed to compile AgeProofCircuit: %v", err)
	}

	_, _, err = groth16.Setup(cs)
	if err != nil {
		t.Fatalf("Failed to setup AgeProofCircuit: %v", err)
	}
}
