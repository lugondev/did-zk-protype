package did

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	cryptomimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	cryptoeddsa "github.com/consensys/gnark-crypto/signature/eddsa"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// DID represents a decentralized identifier
type DID struct {
	ID        string
	PublicKey eddsa.PublicKey
	Document  DIDDocument
}

// DIDDocument stores identity information
type DIDDocument struct {
	Context        []string
	ID             string
	Controller     string
	Authentication []Authentication
	Credentials    []VerifiableCredential
}

// Authentication represents an authentication method
type Authentication struct {
	ID           string
	Type         string
	Controller   string
	PublicKeyJwk map[string]interface{}
}

// VerifiableCredential is credential information issued to an identity
type VerifiableCredential struct {
	Context      []string
	ID           string
	Type         []string
	Issuer       string
	Subject      string
	Claims       map[string]interface{}
	Proof        CredentialProof
	CommitmentID string
}

// CredentialProof stores proof of credential authenticity
type CredentialProof struct {
	Type               string
	Created            string
	VerificationMethod string
	ProofValue         []byte
	ZKProof            []byte
}

// DIDAuthCircuit defines a ZK circuit for DID authentication
type DIDAuthCircuit struct {
	// Public inputs
	PublicKey eddsa.PublicKey `gnark:",public"`
	Message   frontend.Variable
	Signature eddsa.Signature `gnark:",public"`
}

func (circuit *DIDAuthCircuit) Define(api frontend.API) error {
	curve, err := twistededwards.NewEdCurve(api, tedwards.BN254)
	if err != nil {
		return fmt.Errorf("failed to initialize Edwards curve: %w", err)
	}

	// Initialize MiMC hash
	mimc, err := mimc.NewMiMC(api)
	if err != nil {
		return fmt.Errorf("failed to initialize MiMC: %w", err)
	}

	// Hash the message for verification
	// mimc.Write(circuit.Message)
	// hashedMessage := mimc.Sum()

	// Verify signature
	err = eddsa.Verify(curve, circuit.Signature, circuit.Message, circuit.PublicKey, &mimc)
	if err != nil {
		return err
	}

	return nil
}

func (circuit *DIDAuthCircuit) Fill(nbPublic, nbSecret int, values <-chan interface{}) error {
	return nil
}

type AgeProofCircuit struct {
	// Public inputs
	AgeThreshold  frontend.Variable `gnark:",public"`
	AgeCommitment frontend.Variable `gnark:",public"`

	// Private inputs (witness)
	ActualAge         frontend.Variable
	AgeCommitmentSalt frontend.Variable
}

func (circuit *AgeProofCircuit) Define(api frontend.API) error {
	api.AssertIsLessOrEqual(circuit.AgeThreshold, circuit.ActualAge)

	hash, err := mimc.NewMiMC(api)
	if err != nil {
		return fmt.Errorf("failed to create hash: %v", err)
	}

	// Hash the age and salt
	hash.Write(circuit.ActualAge)
	hash.Write(circuit.AgeCommitmentSalt)
	computedCommitment := hash.Sum()

	// Verify the commitment
	api.AssertIsEqual(computedCommitment, circuit.AgeCommitment)

	return nil
}

func (circuit *AgeProofCircuit) Fill(nbPublic, nbSecret int, values <-chan interface{}) error {
	return nil // TODO: Implement proper JSON deserialization if needed
}

// DIDService manages DID-related functionality
type DIDService struct {
	DIDs             map[string]DID
	authCompiled     constraint.ConstraintSystem
	authPk           groth16.ProvingKey
	authVk           groth16.VerifyingKey
	ageProofCompiled constraint.ConstraintSystem
	ageProofPk       groth16.ProvingKey
	ageProofVk       groth16.VerifyingKey
}

// NewDIDService creates a new DID service
func NewDIDService() (*DIDService, error) {
	service := &DIDService{
		DIDs: make(map[string]DID),
	}

	// Compile authentication circuit
	authCompiled, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &DIDAuthCircuit{})
	if err != nil {
		return nil, fmt.Errorf("failed to compile authentication circuit: %v", err)
	}
	service.authCompiled = authCompiled
	fmt.Printf("Authentication circuit compiled\n")

	// Setup parameters for authentication circuit
	authPk, authVk, err := groth16.Setup(authCompiled)
	if err != nil {
		return nil, fmt.Errorf("failed to setup authentication parameters: %v", err)
	}
	service.authPk = authPk
	service.authVk = authVk
	fmt.Printf("Authentication parameters setup complete\n")

	// Compile age proof circuit
	var ageProofCircuit AgeProofCircuit
	ageProofCompiled, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &ageProofCircuit)
	if err != nil {
		return nil, fmt.Errorf("failed to compile age proof circuit: %v", err)
	}
	service.ageProofCompiled = ageProofCompiled

	// Setup parameters for age proof circuit
	ageProofPk, ageProofVk, err := groth16.Setup(ageProofCompiled)
	if err != nil {
		return nil, fmt.Errorf("failed to setup age proof parameters: %v", err)
	}
	service.ageProofPk = ageProofPk
	service.ageProofVk = ageProofVk

	return service, nil
}

// CreateDID creates a new DID
func (s *DIDService) CreateDID() (*DID, *big.Int, error) {

	// Generate private key as big.Int
	privateKey, err := rand.Int(rand.Reader, fr.Modulus())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Convert big.Int to byte array for seed
	seed := make([]byte, 32)
	privKeyBytes := privateKey.Bytes()
	copy(seed[32-len(privKeyBytes):], privKeyBytes)

	privKey, err := cryptoeddsa.New(tedwards.BN254, bytes.NewReader(seed))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create EdDSA private key: %v", err)
	}

	pubKey := privKey.Public()
	pubKeyBytes := pubKey.Bytes()

	id := fmt.Sprintf("did:example:%x", pubKeyBytes)
	fmt.Printf("Created DID: %s\n", id)
	fmt.Printf("CreateDID private key: %x\n", privKey.Bytes())

	// Create DID Document
	document := DIDDocument{
		Context:    []string{"https://www.w3.org/ns/did/v1"},
		ID:         id,
		Controller: id,
		Authentication: []Authentication{
			{
				ID:         id + "#keys-1",
				Type:       "Ed25519VerificationKey2020",
				Controller: id,
				PublicKeyJwk: map[string]interface{}{
					"kty": "OKP",
					"crv": "Ed25519",
					"x":   fmt.Sprintf("%x", pubKeyBytes),
				},
			},
		},
		Credentials: []VerifiableCredential{},
	}

	// Convert to gnark public key format
	gnarkPubKey := eddsa.PublicKey{}
	gnarkPubKey.Assign(tedwards.BN254, pubKeyBytes)
	fmt.Printf("gnarkPubKey public key: %x\n", gnarkPubKey)

	// Save DID
	did := DID{
		ID:        id,
		PublicKey: gnarkPubKey,
		Document:  document,
	}
	s.DIDs[id] = did

	return &did, privateKey, nil
}

// AuthenticateDID authenticates a DID
func (s *DIDService) AuthenticateDID(didID string, privateKey *big.Int, challenge string) ([]byte, []byte, error) {
	did, ok := s.DIDs[didID]
	if !ok {
		return nil, nil, fmt.Errorf("DID does not exist: %s", didID)
	}
	fmt.Printf("Authenticating DID: %s\n", didID)

	// Recreate EdDSA key from private key
	seed := make([]byte, 32)
	privKeyBytes := privateKey.Bytes()
	copy(seed[32-len(privKeyBytes):], privKeyBytes)

	privKey, err := cryptoeddsa.New(tedwards.BN254, bytes.NewReader(seed))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create EdDSA private key: %v", err)
	}

	fmt.Printf("AuthenticateDID private key: %x\n", privKey.Bytes())
	pubKey := privKey.Public()
	pubKeyBytes := pubKey.Bytes()
	gnarkPubKey := eddsa.PublicKey{}
	gnarkPubKey.Assign(tedwards.BN254, pubKeyBytes)
	fmt.Printf("gnarkPubKey public key: %x\n", gnarkPubKey)

	// Hash the challenge to get a numeric value
	hasher := cryptomimc.NewMiMC()
	// hasher.Write([]byte(challenge))
	// hashBytes := hasher.Sum(nil)
	// msgBigInt := new(big.Int).SetBytes(hashBytes)
	// fmt.Printf("Message: %s\n", msgBigInt.String())

	// Sign the message
	signature, err := privKey.Sign([]byte(challenge), hasher)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign challenge: %v", err)
	}
	fmt.Printf("Signature: %x len: %d\n", signature, len(signature))

	gnarkSignature := eddsa.Signature{}
	gnarkSignature.Assign(tedwards.BN254, signature)

	// Debug logging
	fmt.Printf("Debug Info:\n")
	fmt.Printf("- Public Key: %x\n", did.PublicKey)

	// Create the witness
	assignment := &DIDAuthCircuit{
		PublicKey: did.PublicKey,
		Message:   []byte(challenge),
		Signature: gnarkSignature,
	}

	fullWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create witness authentication : %v", err)
	}

	// Generate Groth16 ZK proof for DID authentication
	fmt.Println("Generating authentication proof...")
	// Generate the proof
	proof, err := groth16.Prove(s.authCompiled, s.authPk, fullWitness)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Groth16 DID authentication proof: %v", err)
	}

	// Convert proof to bytes using a buffer
	var buf bytes.Buffer
	if _, err := proof.WriteTo(&buf); err != nil {
		return nil, nil, fmt.Errorf("failed to serialize proof: %v", err)
	}

	return buf.Bytes(), signature, nil
}

// VerifyAuthentication verifies authentication proof
func (s *DIDService) VerifyAuthentication(didID string, proofBytes []byte, signature []byte) (bool, error) {
	did, ok := s.DIDs[didID]
	if !ok {
		return false, fmt.Errorf("DID does not exist: %s", didID)
	}

	// Recover proof from bytes
	g16Proof := groth16.NewProof(ecc.BN254)
	buf := bytes.NewBuffer(proofBytes)
	_, err := g16Proof.ReadFrom(buf)
	if err != nil {
		return false, fmt.Errorf("failed to read proof: %v", err)
	}

	// Hash the challenge to get a numeric value
	// hasher := cryptomimc.NewMiMC()
	// hasher.Write([]byte(challenge))
	// hashBytes := hasher.Sum(nil)
	// msgBigInt := new(big.Int).SetBytes(hashBytes)

	// Create public witness object
	assignment := &DIDAuthCircuit{
		PublicKey: did.PublicKey,
	}
	assignment.Signature.Assign(tedwards.BN254, signature)

	publicWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	if err != nil {
		return false, fmt.Errorf("failed to create witness verify authentication: %v", err)
	}

	// Verify proof
	err = groth16.Verify(g16Proof, s.authVk, publicWitness)
	if err != nil {
		return false, err
	}

	return true, nil
}

// IssueAgeCredential issues age credential
func (s *DIDService) IssueAgeCredential(didID string, age int) (*VerifiableCredential, *big.Int, error) {
	// Generate random salt
	salt, err := rand.Int(rand.Reader, fr.Modulus())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	// Create age commitment using MiMC
	// Instead of using simple multiplication
	hFunc := cryptomimc.NewMiMC()
	ageValue := big.NewInt(int64(age))
	hFunc.Write(ageValue.Bytes())
	hFunc.Write(salt.Bytes())
	commitmentBytes := hFunc.Sum([]byte{})

	// Create credential
	credential := VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1",
		},
		ID:      fmt.Sprintf("%s#credential-1", didID),
		Type:    []string{"VerifiableCredential", "AgeCredential"},
		Issuer:  "did:example:issuer",
		Subject: didID,
		Claims: map[string]interface{}{
			"ageCommitment": fmt.Sprintf("%x", commitmentBytes),
		},
		Proof: CredentialProof{
			Type:               "Ed25519Signature2020",
			Created:            "2023-01-01T00:00:00Z",
			VerificationMethod: "did:example:issuer#keys-1",
			ProofValue:         []byte("example-signature"),
		},
		// Save commitment as hex string
		CommitmentID: fmt.Sprintf("%x", commitmentBytes),
	}

	// Save credential to DID document
	did := s.DIDs[didID]
	did.Document.Credentials = append(did.Document.Credentials, credential)
	s.DIDs[didID] = did

	return &credential, salt, nil
}

// CreateAgeProof creates age proof
func (s *DIDService) CreateAgeProof(didID string, credentialID string, ageThreshold int, actualAge int, salt *big.Int) ([]byte, error) {
	did, ok := s.DIDs[didID]
	if !ok {
		return nil, fmt.Errorf("DID does not exist: %s", didID)
	}

	// Find credential
	var credential *VerifiableCredential
	for i, cred := range did.Document.Credentials {
		if cred.ID == credentialID {
			credential = &did.Document.Credentials[i]
			break
		}
	}
	if credential == nil {
		return nil, fmt.Errorf("credential does not exist: %s", credentialID)
	}

	// Convert commitment from hex string to big.Int
	var commitmentBytes []byte
	n, err := fmt.Sscanf(credential.CommitmentID, "%x", &commitmentBytes)
	if err != nil || n != 1 {
		return nil, fmt.Errorf("failed to parse commitment ID: %v", err)
	}
	commitment := new(big.Int).SetBytes(commitmentBytes)
	if commitment == nil {
		return nil, fmt.Errorf("failed to convert commitment bytes to big.Int")
	}

	// Create witness
	assignment := &AgeProofCircuit{
		AgeThreshold:      ageThreshold,
		AgeCommitment:     commitment,
		ActualAge:         actualAge,
		AgeCommitmentSalt: salt,
	}
	ageWitness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("failed to create witness create age: %v", err)
	}

	// Generate proof
	proof, err := groth16.Prove(s.ageProofCompiled, s.ageProofPk, ageWitness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %v", err)
	}

	// Convert proof to bytes using a buffer
	var buf bytes.Buffer
	if _, err := proof.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("failed to write proof: %v", err)
	}

	return buf.Bytes(), nil
}

// VerifyAgeProof verifies age proof
func (s *DIDService) VerifyAgeProof(didID string, credentialID string, ageThreshold int, proofBytes []byte) (bool, error) {
	did, ok := s.DIDs[didID]
	if !ok {
		return false, fmt.Errorf("DID does not exist: %s", didID)
	}

	// Find credential
	var credential *VerifiableCredential
	for i, cred := range did.Document.Credentials {
		if cred.ID == credentialID {
			credential = &did.Document.Credentials[i]
			break
		}
	}
	if credential == nil {
		return false, fmt.Errorf("credential does not exist: %s", credentialID)
	}

	// Parse commitment ID from hex
	commitmentBytes := make([]byte, len(credential.CommitmentID)/2)
	var n int
	n, err := fmt.Sscanf(credential.CommitmentID, "%x", &commitmentBytes)
	if err != nil || n != 1 {
		return false, fmt.Errorf("failed to parse commitment ID: %v", err)
	}
	commitment := new(big.Int).SetBytes(commitmentBytes)
	if commitment == nil {
		return false, fmt.Errorf("failed to convert commitment bytes to big.Int")
	}

	// Recover proof from bytes
	g16Proof := groth16.NewProof(ecc.BN254)
	buf := bytes.NewBuffer(proofBytes)
	_, err = g16Proof.ReadFrom(buf)
	if err != nil {
		return false, fmt.Errorf("failed to read proof: %v", err)
	}

	// Create public witness object
	assignment := &AgeProofCircuit{
		AgeThreshold:  ageThreshold,
		AgeCommitment: commitment,
	}
	publicWitness, witnessErr := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if witnessErr != nil {
		return false, fmt.Errorf("failed to create witness verify age: %v", witnessErr)
	}

	// Verify proof
	verifyErr := groth16.Verify(g16Proof, s.ageProofVk, publicWitness)
	if verifyErr != nil {
		return false, nil
	}

	return true, nil
}
