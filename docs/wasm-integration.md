# WebAssembly Integration Guide

This document explains the WebAssembly (WASM) integration that replaces JavaScript cryptographic libraries with native Go implementations.

## Overview

The integration moves cryptographic operations from server-side API calls to client-side WebAssembly modules, providing:

- **Zero network latency** for crypto operations
- **Enhanced security** with native Go crypto libraries
- **Better performance** through compiled code
- **Consistent crypto** between client and server
- **Offline capabilities** for cryptographic operations

## Architecture

### Before Integration
```
Client (JavaScript) → HTTP API → Go Backend → Crypto Operations
```

### After Integration
```
Client (WASM Modules) → Direct Crypto Operations
```

## Files Added/Modified

### New Files Created

1. **`client/src/lib/wasm-loader.ts`**
   - WASM module loader and interface definitions
   - Handles loading of crypto.wasm and did.wasm
   - Provides Promise-based API for all WASM functions

2. **`client/src/lib/crypto-wasm.ts`**
   - High-level crypto utilities using WASM
   - Replaces tweetnacl and other JS crypto libraries
   - Provides familiar API for cryptographic operations

3. **`client/src/app/wasm-demo/page.tsx`**
   - Demonstration page showing WASM integration
   - Live examples of crypto and DID operations
   - Before/after comparisons

4. **`client/public/crypto.wasm`** (generated)
   - Compiled Go crypto module
   - Contains: key generation, signing, verification, hashing

5. **`client/public/did.wasm`** (generated)
   - Compiled Go DID module
   - Contains: DID creation, authentication, proof generation

6. **`client/public/wasm_exec.js`** (copied from Go)
   - Go WASM runtime support
   - Required for executing Go WASM modules

### Modified Files

1. **`client/src/app/api/did.ts`**
   - Replaced HTTP API calls with WASM function calls
   - Updated interfaces to match WASM return types
   - Maintains same external API for components

2. **`client/src/components/DIDForm.tsx`**
   - Added WASM module status monitoring
   - Updated UI to show WASM integration status
   - Added loading states for WASM initialization

3. **`client/next.config.ts`**
   - Added WASM support configuration
   - Set proper headers for WASM files
   - Enabled experimental features

4. **`Makefile`** (root)
   - Added WASM build targets
   - Simplified development workflow
   - Added testing and setup commands

## WASM Module Functions

### Crypto Module (`crypto.wasm`)

#### Exported Functions:
- `generateKeyPair()` → `{privateKey: string, publicKey: string}`
- `signMessage(privateKey: string, message: string)` → `{signature: string}`
- `verifySignature(publicKey: string, message: string, signature: string)` → `{verified: boolean}`
- `generateRandomBigInt()` → `{value: string}`
- `hashMessage(message: string)` → `{hash: string}`

#### Usage Example:
```typescript
import { wasmLoader } from '@/lib/wasm-loader';

// Generate key pair
const keyPair = await wasmLoader.generateKeyPair();

// Sign message
const signature = await wasmLoader.signMessage(
  keyPair.privateKey, 
  "Hello, WASM!"
);

// Verify signature
const result = await wasmLoader.verifySignature(
  keyPair.publicKey, 
  "Hello, WASM!", 
  signature.signature
);
```

### DID Module (`did.wasm`)

#### Exported Functions:
- `createDID()` → `{did: {id, publicKey, document}, privateKey: string}`
- `authenticateDID(didID, privateKey, challenge)` → `{proof: string, signature: string}`
- `verifyAuthentication(didID, proof, signature)` → `{verified: boolean}`
- `issueAgeCredential(didID, age)` → `{credential: string, salt: string}`
- `createAgeProof(didID, credentialID, ageThreshold, actualAge, salt)` → `{proof: string}`
- `verifyAgeProof(didID, credentialID, ageThreshold, proof)` → `{verified: boolean}`
- `createMembershipAndBalanceProof(...)` → `{proof: string}`

#### Usage Example:
```typescript
import { wasmLoader } from '@/lib/wasm-loader';

// Create DID
const didResult = await wasmLoader.createDID();

// Authenticate with challenge
const authResult = await wasmLoader.authenticateDID(
  didResult.did.id,
  didResult.privateKey,
  "challenge-string"
);

// Verify authentication
const verified = await wasmLoader.verifyAuthentication(
  didResult.did.id,
  authResult.proof,
  authResult.signature
);
```

## Development Workflow

### Building WASM Modules

```bash
# Build both WASM modules
make build-wasm

# Clean WASM artifacts
make clean-wasm

# Rebuild from scratch
make rebuild-wasm
```

### Development Setup

```bash
# Initial setup
make setup

# Start client development server
make dev-client

# Start backend development server (if needed)
make dev-backend
```

### Testing Integration

1. Build WASM modules: `make build-wasm`
2. Start client: `make dev-client`
3. Visit: `http://localhost:3000/wasm-demo`
4. Run live demonstrations

## API Migration

### Before (HTTP API)
```typescript
// Old approach
const response = await fetch('/api/did/create', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(formData)
});
const result = await response.json();
```

### After (WASM)
```typescript
// New approach
import { wasmLoader } from '@/lib/wasm-loader';
const result = await wasmLoader.createDID();
```

## Performance Benefits

### Network Latency Elimination
- **Before**: Each crypto operation = HTTP round trip (~50-200ms)
- **After**: Direct WASM call (~1-5ms)

### Batch Operations
- **Before**: Multiple API calls for complex workflows
- **After**: Single function calls with internal batching

### Offline Capability
- **Before**: Requires network connection for all operations
- **After**: Full crypto operations work offline

## Security Benefits

### Consistent Cryptography
- Same Go crypto libraries on client and server
- Eliminates potential inconsistencies between implementations
- Uses gnark-crypto for zero-knowledge proof compatibility

### Client-Side Key Management
- Private keys never leave the client
- Reduced attack surface on server
- Enhanced user privacy

## Browser Compatibility

### Requirements
- WebAssembly support (available in all modern browsers)
- JavaScript enabled
- Sufficient memory for WASM modules (~2-5MB)

### Supported Browsers
- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+

## Troubleshooting

### Common Issues

1. **WASM files not loading**
   - Check Next.js configuration in `next.config.ts`
   - Verify WASM files exist in `client/public/`
   - Check browser console for CORS errors

2. **Go runtime errors**
   - Ensure `wasm_exec.js` is from correct Go version
   - Check for JavaScript conflicts
   - Verify WASM module compilation

3. **Memory issues**
   - WASM modules require substantial memory
   - Consider lazy loading for large applications
   - Monitor browser memory usage

### Debugging

```typescript
// Enable WASM debugging
window.addEventListener('error', (e) => {
  console.error('WASM Error:', e);
});

// Check WASM module status
console.log('Crypto ready:', window.wasmCryptoReady);
console.log('DID ready:', window.wasmDIDReady);
```

## Future Enhancements

### Planned Features
1. **Progressive loading** - Load WASM modules on demand
2. **Caching** - Cache compiled WASM modules
3. **Worker threads** - Run WASM in web workers for better performance
4. **Streaming compilation** - Compile WASM while downloading

### Performance Optimizations
1. **Module splitting** - Separate WASM modules by functionality
2. **Compression** - Compress WASM files
3. **Preloading** - Preload WASM modules during app initialization

## Conclusion

The WASM integration provides significant improvements in:
- **Performance**: Eliminated network latency for crypto operations
- **Security**: Client-side key management and consistent crypto
- **User Experience**: Faster operations and offline capability
- **Development**: Simplified API and better error handling

The integration maintains backward compatibility while providing a foundation for advanced cryptographic features and zero-knowledge proofs.