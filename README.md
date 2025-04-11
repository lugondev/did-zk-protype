# DID Example with Zero-Knowledge Proofs

A full-stack example implementation of a Decentralized Identity (DID) system using zero-knowledge proofs and EdDSA signatures. This project demonstrates secure DID creation, age credential issuance, and privacy-preserving authentication.

## Features

- 🔐 Decentralized Identity (DID) creation and management
- 🎭 Zero-knowledge proofs for private age verification
- ✍️ EdDSA signatures for secure authentication
- 🔒 Privacy-preserving credential verification
- 🌐 Modern web interface with real-time DID operations

## Technology Stack

### Backend
- Go HTTP server
- [gnark](https://github.com/consensys/gnark) for zero-knowledge proofs
- Groth16 proving system
- EdDSA on BN254 curve
- In-memory DID storage

### Frontend
- Next.js 15.3.0
- React 19
- TypeScript
- TailwindCSS
- Noble Curves
- TweetNaCl for cryptographic operations

## Getting Started

### Prerequisites
- Go 1.21 or later
- Node.js 20 or later
- pnpm (recommended) or npm

### Setup & Running

1. Start the backend server:
```bash
cd backend
go run main.go
```
The server will start on http://localhost:8080

2. Start the frontend development server:
```bash
cd frontend
pnpm install
pnpm dev
```
The frontend will be available at http://localhost:3000

## API Endpoints

The backend provides three main API endpoints:

### 1. Create DID
- **POST** `/api/did/create`
- Creates a new DID with age credentials

### 2. Authenticate DID
- **POST** `/api/did/authenticate`
- Authenticates a DID using ZKP and EdDSA signatures

### 3. Verify Authentication
- **POST** `/api/did/verify`
- Verifies DID authentication proofs and signatures

For detailed API documentation and request/response formats, see [backend/docs.md](backend/docs.md).

## Development

### Backend Commands
```bash
cd backend
go run main.go     # Start the server
go test ./...      # Run tests
```

### Frontend Commands
```bash
cd frontend
pnpm dev          # Start development server
pnpm build        # Build for production
pnpm start        # Start production server
pnpm lint         # Run linting
```

## Security Features

- Zero-knowledge proofs for privacy-preserving age verification
- EdDSA signatures for secure authentication
- Challenge-response mechanism to prevent replay attacks
- CORS protection
- Input validation and sanitization

## Project Structure

```
.
├── backend/
│   ├── main.go                 # HTTP server implementation
│   ├── docs.md                 # API documentation
│   └── modules/
│       ├── did/               # DID operations
│       ├── eddsa/            # EdDSA signature operations
│       └── rollup/           # ZK rollup implementation
└── frontend/
    ├── src/
    │   ├── app/              # Next.js app directory
    │   ├── components/       # React components
    │   └── types/           # TypeScript definitions
    └── public/              # Static assets
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open-source software.
