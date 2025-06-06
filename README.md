# gnark DID Example

This project demonstrates a Decentralized Identity (DID) system using zero-knowledge proofs with the gnark library. It consists of a Go backend with cryptographic operations and a Next.js frontend that can utilize WebAssembly modules for client-side cryptographic operations.

## Project Structure

```
.
├── backend/                 # Go backend with DID operations
│   ├── cmd/
│   │   ├── api/            # REST API server
│   │   └── wasm/           # WebAssembly entry points
│   │       ├── did/        # DID operations WASM module
│   │       └── crypto/     # Cryptographic utilities WASM module
│   ├── pkg/
│   │   └── did/            # Core DID functionality
│   └── internal/           # Internal services and handlers
├── client/                 # Next.js frontend
│   ├── pkg/               # Generated WASM modules (created by build)
│   └── src/               # Frontend source code
└── Makefile               # Build automation for WASM modules
```

## Features

### Backend (Go)
- **DID Creation**: Generate decentralized identifiers with EdDSA key pairs
- **Authentication**: Zero-knowledge proof-based DID authentication
- **Age Credentials**: Issue and verify age credentials with ZK proofs
- **Membership Proofs**: Create proofs of organization membership and balance ranges
- **REST API**: HTTP endpoints for DID operations

### Frontend (Next.js)
- **User Interface**: Modern React-based interface for DID management
- **Client-side Crypto**: WebAssembly modules for browser-based cryptographic operations
- **Integration**: Seamless integration with backend services

### WebAssembly Modules
- **DID Module** (`client/pkg/did.wasm`): Client-side DID operations
- **Crypto Module** (`client/pkg/crypto.wasm`): Low-level cryptographic functions

## Quick Start

### Prerequisites
- Go 1.19 or later
- Node.js 18 or later
- pnpm (for client dependencies)

### Build WebAssembly Modules

```bash
# Build WASM modules for the client
make build-wasm

# Or for development with verbose output
make build-wasm-dev

# Or for production with optimizations
make build-wasm-prod
```

### Available Make Targets

```bash
make help                    # Show all available targets
make check-deps             # Verify required dependencies
make install-deps           # Install Go dependencies
make build-wasm             # Build WASM modules
make build-wasm-dev         # Development build with verbose output
make build-wasm-prod        # Production build with optimizations
make clean                  # Clean build artifacts
make verify-wasm           # Verify built WASM modules
make info-wasm             # Show WASM module information
make all                   # Complete build pipeline
```

### Run the Backend

```bash
cd backend
go run cmd/api/main.go
```

### Run the Frontend

```bash
cd client
pnpm install
pnpm dev
```

## WebAssembly Integration

### Loading WASM Modules in JavaScript

```javascript
// Load the Go WASM runtime
import './pkg/wasm_exec.js';

// Load DID operations module
const didWasm = await WebAssembly.instantiateStreaming(
  fetch('/pkg/did.wasm'),
  go.importObject
);

// Load crypto utilities module
const cryptoWasm = await WebAssembly.instantiateStreaming(
  fetch('/pkg/crypto.wasm'),
  go.importObject
);

// Wait for modules to be ready
while (!globalThis.wasmDIDReady || !globalThis.wasmCryptoReady) {
  await new Promise(resolve => setTimeout(resolve, 100));
}
```

### Available WASM Functions

#### DID Operations (`did.wasm`)
- `createDID()` - Create a new DID with key pair
- `authenticateDID(didID, privateKey, challenge)` - Authenticate a DID
- `verifyAuthentication(didID, proof, signature)` - Verify authentication proof
- `issueAgeCredential(didID, age)` - Issue age credential
- `createAgeProof(didID, credentialID, ageThreshold, actualAge, salt)` - Create age proof
- `verifyAgeProof(didID, credentialID, ageThreshold, proof)` - Verify age proof
- `createMembershipAndBalanceProof(orgID, balance, min, max, salt)` - Create membership proof

#### Crypto Operations (`crypto.wasm`)
- `generateKeyPair()` - Generate EdDSA key pair
- `signMessage(privateKey, message)` - Sign a message
- `verifySignature(publicKey, message, signature)` - Verify signature
- `generateRandomBigInt()` - Generate random big integer
- `hashMessage(message)` - Hash a message

### Example Usage

```javascript
// Create a new DID
const didResult = await createDID();
const { did, privateKey } = JSON.parse(didResult);

// Authenticate the DID
const challenge = "random_challenge_string";
const authResult = await authenticateDID(did.id, privateKey, challenge);
const { proof, signature } = JSON.parse(authResult);

// Verify authentication
const verifyResult = await verifyAuthentication(did.id, proof, signature);
const { verified } = JSON.parse(verifyResult);
console.log('Authentication verified:', verified);
```

## API Endpoints

The backend provides RESTful API endpoints:

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - User login
- `POST /api/auth/did-login` - DID-based login
- `POST /api/users/profile` - Get user profile
- `PUT /api/users/profile` - Update profile
- `POST /api/did/create` - Create DID
- `POST /api/did/authenticate` - DID authentication
- `POST /api/did/verify` - Verify DID proof

## Development

### Building for Development
```bash
# Build WASM with debug information
make build-wasm-dev

# Watch for changes and rebuild
make clean && make build-wasm-dev
```

### Building for Production
```bash
# Build optimized WASM modules
make build-wasm-prod
```

### Testing WASM Modules
```bash
# Verify modules are built correctly
make verify-wasm

# Show module information
make info-wasm
```

## Architecture

The system uses:
- **gnark**: Zero-knowledge proof framework
- **EdDSA**: Digital signatures on twisted Edwards curves
- **Groth16**: Zero-knowledge proof system
- **BN254**: Elliptic curve for cryptographic operations
- **MiMC**: Hash function for commitments

## Security Considerations

- Private keys are generated using cryptographically secure random number generation
- Zero-knowledge proofs ensure privacy-preserving authentication
- Age credentials use commitments to hide actual age while proving thresholds
- All cryptographic operations use well-established algorithms and libraries

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test the WASM build: `make clean && make all`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
