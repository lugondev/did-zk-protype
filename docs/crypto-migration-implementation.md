# Cryptographic Operations Migration: Backend to Client-Side WASM

## Overview

This document details the implementation of moving cryptographic operations from the backend to client-side using WebAssembly (WASM) modules. The migration improves security, reduces server load, and provides better user privacy.

## 1. Current Cryptographic Operations Identified

### Backend Operations (Before Migration)

#### Authentication Service (`backend/internal/services/auth_service.go`)
- **Password Hashing**: Using `bcrypt.GenerateFromPassword()` and `bcrypt.CompareHashAndPassword()`
- **JWT Token Generation**: Using `jwt.NewWithClaims()` and `token.SignedString()`
- **DID Authentication**: Using `didService.VerifyAuthentication()` with hex decoding

#### DID Service (`backend/internal/did/handlers.go`)
- **DID Creation**: `didService.CreateDID()` 
- **DID Authentication**: `didService.AuthenticateDID()` with private key and challenge
- **Signature Verification**: `didService.VerifyAuthentication()` with proof and signature
- **Age Credential Issuance**: `didService.IssueAgeCredential()`
- **Zero-Knowledge Proof Generation**: `didService.CreateMembershipAndBalanceProof()`

### WASM Modules Available

#### Crypto WASM (`backend/cmd/wasm/crypto/main.go`)
- `generateKeyPair()`: EdDSA key pair generation
- `signMessage()`: Message signing with private key
- `verifySignature()`: Signature verification
- `generateRandomBigInt()`: Cryptographically secure random number generation
- `hashMessage()`: Message hashing

#### DID WASM (`backend/cmd/wasm/did/main.go`)
- `createDID()`: Complete DID creation with cryptographic keys
- `authenticateDID()`: DID authentication with challenge-response
- `verifyAuthentication()`: Client-side verification of DID proofs
- `issueAgeCredential()`: Age credential issuance
- `createAgeProof()`: Zero-knowledge age proof generation
- `createMembershipAndBalanceProof()`: Membership and balance range proofs

## 2. Migration Implementation

### Enhanced Client-Side Authentication Library

Created `client/src/lib/wasm-auth.ts` with the following capabilities:

```typescript
export class WASMAuth {
  // Password hashing with client-side salt generation
  async hashPassword(password: string, salt?: string): Promise<PasswordHashResult>
  
  // Authentication keypair generation
  async generateAuthKeyPair(): Promise<{privateKey: string; publicKey: string}>
  
  // Complete DID creation with credentials
  async createDIDClientSide(name: string, dob: string): Promise<DIDCreationResult>
  
  // DID authentication with immediate verification
  async authenticateDIDClientSide(didId: string, privateKey: string, challenge: string)
  
  // Zero-knowledge proof generation
  async createMembershipProofClientSide(orgId: string, balance: number, min: number, max: number)
  
  // Enhanced login with client-side crypto
  async loginWithClientSideCrypto(username: string, password: string): Promise<AuthResponse>
  
  // Enhanced DID login with full client-side verification
  async loginWithDIDClientSide(didId: string, privateKey: string): Promise<AuthResponse>
}
```

### Enhanced UI Components

#### Enhanced Login Form (`client/src/components/EnhancedLoginForm.tsx`)
- **Crypto Status Indicators**: Real-time display of cryptographic operation progress
- **WASM Authentication**: Button to use client-side crypto instead of backend
- **Performance Monitoring**: Shows client-side operation timing

#### Enhanced DID Form (`client/src/components/EnhancedDIDForm.tsx`)
- **Client-Side DID Creation**: Complete DID creation without backend calls
- **Membership Proof Generation**: Zero-knowledge proofs generated client-side
- **Crypto Testing**: Built-in testing for all WASM operations

#### Migration Demo (`client/src/app/crypto-migration-demo/page.tsx`)
- **Side-by-Side Comparison**: Backend vs WASM performance comparison
- **Operation Timing**: Precise measurement of crypto operation performance
- **Result Visualization**: Clear display of migration benefits

## 3. Code Examples

### Before: Backend Password Authentication

```typescript
// client/src/lib/auth.ts (Original)
export async function login(data: LoginData): Promise<AuthResponse> {
  const response = await fetch(buildApiUrl('/auth/login'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data), // Plain password sent to backend
  });
  // Backend performs bcrypt hashing and verification
}
```

```go
// backend/internal/services/auth_service.go (Original)
func (s *authService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
    user, err := s.userRepo.GetByUsername(ctx, req.Username)
    // Password verification happens on server
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        return nil, ErrInvalidCredentials
    }
}
```

### After: Client-Side WASM Authentication

```typescript
// client/src/lib/wasm-auth.ts (Enhanced)
async loginWithClientSideCrypto(username: string, password: string): Promise<AuthResponse> {
  // Generate client-side password hash
  const { hash: passwordHash } = await this.hashPassword(password);
  
  // Generate authentication keypair
  const { publicKey } = await this.generateAuthKeyPair();
  
  // Send pre-hashed data to backend
  const authData: ClientSideAuthData = {
    username,
    password: '', // No plain password
    clientSideHash: passwordHash,
    publicKey,
  };
  
  const response = await fetch(buildApiUrl('/auth/login-enhanced'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(authData),
  });
}
```

### Before: Backend DID Authentication

```typescript
// client/src/app/api/did.ts (Original - partially migrated)
export async function authenticateDID(didId: string, privateKey: string, challenge: string) {
  // Still calls backend API for DID authentication
  const result = await wasmLoader.authenticateDID(didId, privateKey, challenge);
  return { proof: result.proof, signature: result.signature };
}
```

### After: Complete Client-Side DID Authentication

```typescript
// client/src/lib/wasm-auth.ts (Enhanced)
async authenticateDIDClientSide(didId: string, privateKey: string, challenge: string) {
  // Authenticate using WASM
  const authResult = await wasmLoader.authenticateDID(didId, privateKey, challenge);
  
  // Verify the authentication immediately client-side
  const verifyResult = await wasmLoader.verifyAuthentication(
    didId, authResult.proof, authResult.signature
  );

  return {
    proof: authResult.proof,
    signature: authResult.signature,
    verified: verifyResult.verified, // Client-side verification result
  };
}
```

### Zero-Knowledge Proof Generation

```typescript
// client/src/lib/wasm-auth.ts
async createMembershipProofClientSide(
  organizationId: string, balance: number, balanceRangeMin: number, balanceRangeMax: number
) {
  // Generate salt for the proof
  const saltBytes = await cryptoWasm.generateChallenge(32);
  const salt = cryptoWasm.uint8ArrayToHex(saltBytes);

  // Create the membership and balance proof
  const proofResult = await wasmLoader.createMembershipAndBalanceProof(
    organizationId, balance, balanceRangeMin, balanceRangeMax, salt
  );

  // Create commitment and organization hash client-side
  const orgIdBytes = cryptoWasm.stringToUint8Array(organizationId);
  const orgIdHash = await cryptoWasm.hashMessage(orgIdBytes);
  
  // Generate commitment (hash of orgId + balance + salt)
  const commitmentData = new Uint8Array(/* combined data */);
  const commitment = await cryptoWasm.hashMessage(commitmentData);

  return {
    proof: proofResult.proof,
    salt,
    commitment: cryptoWasm.uint8ArrayToHex(commitment),
    organizationIdHash: cryptoWasm.uint8ArrayToHex(orgIdHash),
  };
}
```

## 4. Migration Benefits

### Security Improvements
- **No Plain Text Passwords**: Passwords are hashed client-side before transmission
- **Private Key Privacy**: Private keys never leave the client
- **Reduced Attack Surface**: Server doesn't handle sensitive cryptographic data
- **Client-Side Verification**: Immediate verification without server round-trips

### Performance Benefits
- **Reduced Server Load**: Cryptographic operations moved to client
- **Faster Response Times**: No network latency for crypto operations
- **Parallel Processing**: Multiple crypto operations can run simultaneously
- **Caching**: WASM modules loaded once and reused

### User Experience Improvements
- **Real-Time Feedback**: Crypto status indicators show operation progress
- **Offline Capability**: Some operations can work without server connectivity
- **Privacy**: Users maintain control over their cryptographic data

## 5. Usage Examples

### Enhanced Login with Crypto Status

```tsx
// client/src/components/EnhancedLoginForm.tsx
const [cryptoStatus, setCryptoStatus] = useState({
  keyGenerated: false,
  hashComputed: false,
  didVerified: false,
});

const handlePasswordLogin = async (e: React.FormEvent) => {
  // Step 1: Generate authentication keypair
  setCryptoStatus(prev => ({ ...prev, keyGenerated: true }));
  
  // Step 2: Perform client-side crypto login
  setCryptoStatus(prev => ({ ...prev, hashComputed: true }));
  await wasmAuth.loginWithClientSideCrypto(formData.username, formData.password);
};

// Visual feedback component
const CryptoStatusIndicator = () => (
  <div className="flex items-center space-x-2">
    <Cpu className="w-3 h-3" />
    <span>Client-side crypto:</span>
    <div className={`w-2 h-2 rounded-full ${cryptoStatus.keyGenerated ? 'bg-green-500' : 'bg-gray-300'}`} />
    <span>Keys</span>
    <div className={`w-2 h-2 rounded-full ${cryptoStatus.hashComputed ? 'bg-green-500' : 'bg-gray-300'}`} />
    <span>Hash</span>
  </div>
);
```

### Migration Comparison Demo

```tsx
// client/src/app/crypto-migration-demo/page.tsx
const runAuthenticationComparison = async () => {
  // Backend Authentication (Traditional)
  const backendStart = performance.now();
  const backendResult = await traditionalLogin(username, password);
  const backendDuration = performance.now() - backendStart;

  // Client-side WASM Authentication
  const wasmStart = performance.now();
  const wasmResult = await wasmAuth.loginWithClientSideCrypto(username, password);
  const wasmDuration = performance.now() - wasmStart;

  // Performance comparison
  const improvement = ((backendDuration - wasmDuration) / backendDuration * 100);
  console.log(`WASM is ${improvement.toFixed(1)}% faster`);
};
```

## 6. Backend Updates Required

To fully utilize the client-side cryptographic operations, the backend should be updated with new endpoints:

### Enhanced Authentication Endpoints

```go
// backend/internal/handlers/auth_handlers.go (Proposed)
func LoginEnhanced(c *fiber.Ctx) error {
    var req struct {
        Username       string `json:"username"`
        ClientSideHash string `json:"clientSideHash"`
        PublicKey      string `json:"publicKey"`
    }
    
    // Verify client-side hash against stored hash
    // No bcrypt comparison needed - client already verified
    // Generate JWT token and return
}

func LoginDIDEnhanced(c *fiber.Ctx) error {
    var req struct {
        DIDID     string `json:"didId"`
        Proof     string `json:"proof"`
        Signature string `json:"signature"`
        Challenge string `json:"challenge"`
        PublicKey string `json:"publicKey"`
    }
    
    // Client has already verified the DID authentication
    // Server just needs to validate the proof format and generate session
}
```

## 7. Testing and Validation

### Crypto Testing Suite

The enhanced components include built-in testing capabilities:

```typescript
const handleTestCrypto = async () => {
  // Test key generation
  const keyPair = await wasmAuth.generateAuthKeyPair();
  
  // Test password hashing
  const hash = await wasmAuth.hashPassword('test-password');
  
  // Test signing and verification
  const signature = await wasmAuth.signDataClientSide(keyPair.privateKey, 'test-data');
  const verified = await wasmAuth.verifySignatureClientSide(keyPair.publicKey, 'test-data', signature);
  
  if (verified) {
    console.log('All crypto operations working correctly!');
  }
};
```

## 8. Migration Status

### ✅ Completed
- Client-side password hashing
- Client-side DID creation and authentication
- Zero-knowledge proof generation
- Enhanced UI components with crypto status indicators
- Performance comparison demo
- WASM integration for all major crypto operations

### 🔄 In Progress
- Backend endpoint updates for enhanced authentication
- Complete migration of all DID operations
- Integration testing with backend

### 📋 Next Steps
1. Update backend authentication endpoints to accept client-side hashed data
2. Implement enhanced DID login endpoints
3. Add comprehensive error handling for WASM operations
4. Performance optimization and caching strategies
5. Security audit of client-side crypto implementation

## 9. Security Considerations

### Client-Side Security
- WASM modules are cryptographically secure and tamper-resistant
- Private keys are generated and stored client-side only
- All cryptographic operations use proven algorithms (EdDSA, MiMC)

### Network Security
- Sensitive data (private keys, plain passwords) never transmitted
- Only public keys and hashed/signed data sent to server
- Reduced attack surface on backend systems

### Implementation Notes
- Always validate client-side crypto results on the backend when necessary
- Implement proper error handling for WASM loading failures
- Use secure random number generation for all cryptographic operations
- Regular security audits of WASM modules and client-side implementations

This migration significantly enhances the security and performance of the cryptographic operations while maintaining compatibility with existing backend systems.