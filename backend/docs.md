# DID Example Backend Documentation

## Overview

This backend implements a Decentralized Identifier (DID) system using zero-knowledge proofs (ZKP) and EdDSA signatures. It provides functionality for creating DIDs, issuing age credentials, and performing secure authentication without revealing sensitive information.

## Architecture

The backend is structured into several components:

### Main Server (`main.go`)
- HTTP server implementation with RESTful API endpoints
- Handles CORS and request/response formatting
- Interfaces with the DID service for core functionality

### DID Module (`modules/did/did.go`)
- Core DID operations implementation
- Manages DID creation, authentication, and verification
- Handles age credentials using zero-knowledge proofs
- Implements EdDSA signature operations

## API Endpoints

### 1. Create DID (`POST /api/did/create`)
Creates a new DID with associated age credentials.

**Request:**
```json
{
  "name": "string",
  "dob": "string" // Date of birth
}
```

**Response:**
```json
{
  "did": {
    "id": "string",
    "publicKey": "object",
    "document": "object"
  },
  "privateKey": "string" // Hex-encoded private key
}
```

### 2. Authenticate DID (`POST /api/did/authenticate`)
Authenticates a DID using ZKP and EdDSA signatures.

**Request:**
```json
{
  "didId": "string",
  "privateKey": "string", // Hex-encoded
  "challenge": "string"
}
```

**Response:**
```json
{
  "proof": "string", // Hex-encoded ZKP
  "signature": "string" // Hex-encoded EdDSA signature
}
```

### 3. Verify Authentication (`POST /api/did/verify`)
Verifies DID authentication proofs and signatures.

**Request:**
```json
{
  "didId": "string",
  "signature": "string", // Hex-encoded
  "proof": "string" // Hex-encoded
}
```

**Response:**
```json
{
  "verified": "boolean"
}
```

## Core Components

### 1. DID Service
- Manages DID lifecycle
- Handles cryptographic operations
- Stores DID information in memory
- Issues and verifies age credentials

### 2. Zero-Knowledge Proofs
- Uses Groth16 proving system
- Implements two main circuits:
  - DID Authentication Circuit
  - Age Proof Circuit
- Provides privacy-preserving verification

### 3. EdDSA Signatures
- Uses BN254 curve for Edwards-curve Digital Signature Algorithm
- Provides cryptographic authentication
- Ensures message integrity and non-repudiation

## Security Features

1. **Zero-Knowledge Proofs**
   - Allows age verification without revealing actual age
   - Provides privacy-preserving authentication
   - Uses Groth16 ZK-SNARK protocol

2. **EdDSA Signatures**
   - Secure digital signatures on Edwards curves
   - Provides authenticity and integrity
   - Non-repudiation of signed messages

3. **Challenge-Response Authentication**
   - Prevents replay attacks
   - Ensures freshness of authentication

## Dependencies

- `github.com/consensys/gnark` - For zero-knowledge proof systems
- `github.com/consensys/gnark-crypto` - For cryptographic operations
- Standard Go libraries for HTTP server and encoding

## Error Handling

The server implements comprehensive error handling:
- Invalid request format
- Authentication failures
- Verification errors
- Internal service errors

All errors are returned with appropriate HTTP status codes and descriptive messages.

## CORS Configuration

The server includes CORS middleware that:
- Allows all origins (`*`)
- Supports `POST`, `GET`, and `OPTIONS` methods
- Accepts `Content-Type` header
- Handles preflight requests

## Best Practices

1. **Input Validation**
   - All request parameters are validated
   - Type checking for all inputs
   - Proper error messages for invalid inputs

2. **Security**
   - No sensitive data in logs
   - Proper error handling
   - Use of cryptographic primitives
   - Stateless authentication

3. **API Design**
   - RESTful endpoints
   - Consistent error formats
   - Clear request/response structures
   - Proper content type headers
