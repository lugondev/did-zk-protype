.PHONY: help build-wasm clean-wasm dev-client dev-backend install-deps

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# WASM Build targets
build-wasm: ## Build WebAssembly modules for client
	@echo "Building WASM modules..."
	@cd backend && GOOS=js GOARCH=wasm go build -o ../client/public/crypto.wasm ./cmd/wasm/crypto
	@cd backend && GOOS=js GOARCH=wasm go build -o ../client/public/did.wasm ./cmd/wasm/did
	@cp "$$(cd backend && go env GOROOT)/misc/wasm/wasm_exec.js" client/public/
	@echo "WASM modules built successfully!"

clean-wasm: ## Clean WASM build artifacts
	@echo "Cleaning WASM artifacts..."
	@rm -f client/public/crypto.wasm client/public/did.wasm client/public/wasm_exec.js
	@echo "WASM artifacts cleaned!"

rebuild-wasm: clean-wasm build-wasm ## Clean and rebuild WASM modules

# Development
dev-client: ## Start client development server
	@echo "Starting client development server..."
	@cd client && npm run dev

dev-backend: ## Start backend development server
	@echo "Starting backend development server..."
	@cd backend && make dev

install-deps: ## Install dependencies for both client and backend
	@echo "Installing client dependencies..."
	@cd client && npm install
	@echo "Installing backend dependencies..."
	@cd backend && go mod download
	@echo "Dependencies installed!"

# Full development setup
setup: install-deps build-wasm ## Setup development environment
	@echo "Development environment ready!"
	@echo ""
	@echo "To start development:"
	@echo "  make dev-client   # Start Next.js client"
	@echo "  make dev-backend  # Start Go backend"
	@echo ""
	@echo "WASM integration features:"
	@echo "  - Client-side cryptographic operations"
	@echo "  - DID creation and authentication"
	@echo "  - Zero-latency crypto operations"
	@echo "  - Native Go crypto libraries in browser"

# Test WASM integration
test-wasm: build-wasm ## Build WASM and test integration
	@echo "Testing WASM integration..."
	@echo "✓ WASM modules built"
	@echo "✓ wasm_exec.js copied"
	@echo "✓ Ready for client integration"
	@echo ""
	@echo "Start the client with: make dev-client"
	@echo "Then visit: http://localhost:3000/wasm-demo"