# DID Example with Zero-Knowledge Proofs

A full-stack example implementation of a Decentralized Identity (DID) system using zero-knowledge proofs and EdDSA signatures. This project demonstrates secure DID creation, age credential issuance, and privacy-preserving authentication.

## Features

- 🔐 Decentralized Identity (DID) creation and management
- 🎭 Zero-knowledge proofs for private age verification
- ✍️ EdDSA signatures for secure authentication
- 🔒 Privacy-preserving credential verification
- 🌐 Modern web interface with real-time DID operations
- 📱 Responsive dashboard for credential management
- 🔑 Client-side cryptographic operations

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
- TailwindCSS with shadcn/ui components
- Noble Curves for cryptographic operations
- TweetNaCl for additional crypto functionality

## Getting Started

### Prerequisites
- Go 1.21 or later
- Node.js 20 or later
- pnpm (recommended) or npm

### Setup & Running

1. Start the backend server:
```bash
cd backend
go mod download
go run main.go
```
The server will start on http://localhost:8080

2. Start the frontend development server:
```bash
cd client
pnpm install
pnpm dev
```
The frontend will be available at http://localhost:3000

## Project Structure

```
.
├── backend/                 # Go backend service
│   ├── main.go             # HTTP server implementation
│   ├── docs.md             # API documentation
│   └── modules/
│       ├── auth/           # Authentication
│       ├── db/             # Database operations
│       ├── did/            # DID operations
│       ├── eddsa/          # EdDSA operations
│       └── rollup/         # ZK rollup implementation
└── client/                 # Next.js frontend
    ├── src/
    │   ├── app/           # Next.js pages and API routes
    │   ├── components/    # React components
    │   ├── lib/          # Utility functions
    │   └── types/        # TypeScript definitions
    └── public/           # Static assets
```

## Documentation

- [Backend Documentation](backend/README.md)
  - API endpoints
  - Module architecture
  - Security features
  - Testing instructions

- [Frontend Documentation](client/README.md)
  - Component structure
  - State management
  - Security considerations
  - Development guidelines

## Core Features

### DID Management
- Create and manage Decentralized Identifiers
- Issue age credentials with privacy preservation
- Authenticate using ZKP and EdDSA signatures

### Zero-Knowledge Proofs
- Age verification without revealing actual age
- Groth16 proving system implementation
- Privacy-preserving authentication

### Security
- Client-side cryptographic operations
- Challenge-response authentication
- Protection against replay attacks
- Secure credential storage
- Input validation and sanitization

## Development

### Backend Commands
```bash
cd backend
go test ./...      # Run all tests
go test ./modules/did  # Test specific module
```

### Frontend Commands
```bash
cd client
pnpm dev          # Start development server
pnpm build        # Build for production
pnpm start        # Start production server
pnpm lint         # Run linting
```

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

For detailed contribution guidelines, see the README files in the respective submodules.

## License

This project is open-source software.
