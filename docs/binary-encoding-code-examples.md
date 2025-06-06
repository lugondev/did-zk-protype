# Binary Encoding Implementation - Code Examples

This document provides concrete code examples of the binary encoding implementation for WebAssembly data transfer.

## 1. Modified WASM Functions in `crypto-wasm.ts`

### Before (String-based)
```typescript
export interface CryptoKeyPair {
    privateKey: string;
    publicKey: string;
}

async generateKeyPair(): Promise<CryptoKeyPair> {
    const result = await wasmLoader.generateKeyPair();
    return {
        privateKey: result.privateKey,
        publicKey: result.publicKey,
    };
}

async signMessage(privateKey: string, message: string): Promise<string> {
    const result = await wasmLoader.signMessage(privateKey, message);
    return result.signature;
}
```

### After (Binary-based)
```typescript
export interface CryptoKeyPair {
    privateKey: Uint8Array;
    publicKey: Uint8Array;
}

async generateKeyPair(): Promise<CryptoKeyPair> {
    const result = await wasmLoader.generateKeyPair();
    return {
        privateKey: this.hexToUint8Array(result.privateKey),
        publicKey: this.hexToUint8Array(result.publicKey),
    };
}

async signMessage(privateKey: Uint8Array, message: Uint8Array): Promise<Uint8Array> {
    const privateKeyHex = this.uint8ArrayToHex(privateKey);
    const messageHex = this.uint8ArrayToHex(message);
    const result = await wasmLoader.signMessage(privateKeyHex, messageHex);
    return this.hexToUint8Array(result.signature);
}
```

### Added Utility Functions
```typescript
/**
 * Convert string to Uint8Array
 */
stringToUint8Array(str: string): Uint8Array {
    return new TextEncoder().encode(str);
}

/**
 * Convert Uint8Array to string
 */
uint8ArrayToString(bytes: Uint8Array): string {
    return new TextDecoder().decode(bytes);
}

/**
 * Convert hex string to Uint8Array
 */
hexToUint8Array(hex: string): Uint8Array {
    const cleanHex = hex.startsWith('0x') ? hex.slice(2) : hex;
    const bytes = new Uint8Array(cleanHex.length / 2);
    for (let i = 0; i < cleanHex.length; i += 2) {
        bytes[i / 2] = parseInt(cleanHex.substr(i, 2), 16);
    }
    return bytes;
}

/**
 * Convert Uint8Array to hex string
 */
uint8ArrayToHex(bytes: Uint8Array): string {
    return Array.from(bytes)
        .map(byte => byte.toString(16).padStart(2, '0'))
        .join('');
}
```

## 2. Enhanced API Functions in `did.ts`

### Binary Authentication Function
```typescript
// Binary version of authenticate DID for optimized data transfer
export async function authenticateDIDBinary(
    didId: string,
    privateKey: Uint8Array,
    challenge: Uint8Array
): Promise<{
    proof: string;
    signature: Uint8Array;
}> {
    try {
        // Convert binary data to hex for WASM call
        const privateKeyHex = cryptoWasm.uint8ArrayToHex(privateKey);
        const challengeHex = cryptoWasm.uint8ArrayToHex(challenge);
        
        // Use WASM instead of API call
        const result = await wasmLoader.authenticateDID(didId, privateKeyHex, challengeHex);
        return {
            proof: result.proof,
            signature: cryptoWasm.hexToUint8Array(result.signature),
        };
    } catch (error) {
        if (error instanceof Error) {
            throw error;
        }
        throw new Error('Failed to authenticate DID with binary encoding');
    }
}
```

### Binary Verification Function
```typescript
// Binary version of verify DID for optimized data transfer
export async function verifyDIDBinary(
    didId: string,
    signature: Uint8Array,
    proof: string
): Promise<boolean> {
    try {
        // Convert binary signature to hex for WASM call
        const signatureHex = cryptoWasm.uint8ArrayToHex(signature);
        
        // Use WASM instead of API call
        const result = await wasmLoader.verifyAuthentication(didId, proof, signatureHex);
        return result.verified;
    } catch (error) {
        throw new Error('Failed to verify DID with binary encoding');
    }
}
```

### DIDCrypto Utility Object
```typescript
// Binary crypto helper functions for DID operations
export const DIDCrypto = {
    /**
     * Generate a cryptographic challenge as binary data
     */
    async generateChallengeBinary(length: number = 32): Promise<Uint8Array> {
        return await cryptoWasm.generateChallenge(length);
    },

    /**
     * Generate a keypair with binary encoding
     */
    async generateKeyPairBinary(): Promise<{
        privateKey: Uint8Array;
        publicKey: Uint8Array;
    }> {
        return await cryptoWasm.generateKeyPair();
    },

    /**
     * Sign a message with binary data
     */
    async signMessageBinary(privateKey: Uint8Array, message: Uint8Array): Promise<Uint8Array> {
        return await cryptoWasm.signMessage(privateKey, message);
    },

    /**
     * Verify a signature with binary data
     */
    async verifySignatureBinary(publicKey: Uint8Array, message: Uint8Array, signature: Uint8Array): Promise<boolean> {
        return await cryptoWasm.verifySignature(publicKey, message, signature);
    },

    /**
     * Hash a message with binary encoding
     */
    async hashMessageBinary(message: Uint8Array): Promise<Uint8Array> {
        return await cryptoWasm.hashMessage(message);
    },

    // Utility functions for conversion
    stringToUint8Array: (str: string) => cryptoWasm.stringToUint8Array(str),
    uint8ArrayToString: (bytes: Uint8Array) => cryptoWasm.uint8ArrayToString(bytes),
    hexToUint8Array: (hex: string) => cryptoWasm.hexToUint8Array(hex),
    uint8ArrayToHex: (bytes: Uint8Array) => cryptoWasm.uint8ArrayToHex(bytes),
};
```

## 3. Updated Component Usage in `DIDForm.tsx`

### Binary Encoding Toggle
```typescript
const [useBinaryEncoding, setUseBinaryEncoding] = useState<boolean>(true);

// UI Toggle
<div className='mb-6 p-4 bg-purple-50 border border-purple-200 rounded-md'>
    <div className='flex items-center justify-between'>
        <div>
            <h3 className='text-sm font-medium text-purple-900'>Binary Data Encoding</h3>
            <p className='text-xs text-purple-700 mt-1'>
                {useBinaryEncoding ? 'Using Uint8Array for optimized data transfer' : 'Using string encoding for compatibility'}
            </p>
        </div>
        <label className='flex items-center cursor-pointer'>
            <input 
                type='checkbox' 
                checked={useBinaryEncoding} 
                onChange={(e) => setUseBinaryEncoding(e.target.checked)} 
                className='sr-only' 
                disabled={useBatching} 
            />
            <div className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${useBinaryEncoding && !useBatching ? 'bg-purple-600' : 'bg-gray-200'} ${useBatching ? 'opacity-50 cursor-not-allowed' : ''}`}>
                <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${useBinaryEncoding && !useBatching ? 'translate-x-6' : 'translate-x-1'}`} />
            </div>
        </label>
    </div>
</div>
```

### Binary Authentication Logic
```typescript
const handleAuthenticate = async () => {
    // ... existing code ...
    
    if (useBinaryEncoding) {
        // Use binary encoding for optimized data transfer
        const privateKeyBytes = DIDCrypto.hexToUint8Array(didInfo.privateKey)
        const challengeBytes = DIDCrypto.stringToUint8Array(authenticationState.challenge)
        
        const response = await authenticateDIDBinary(didInfo.did.ID, privateKeyBytes, challengeBytes)
        console.log('Binary authentication response:', response)

        setAuthenticationState((prev) => ({
            ...prev,
            proof: response.proof,
            signature: DIDCrypto.uint8ArrayToHex(response.signature),
        }))
    } else {
        // Use individual authentication operation with string encoding
        const response = await authenticateDID(didInfo.did.ID, didInfo.privateKey, authenticationState.challenge)
        // ... handle response ...
    }
};
```

### Binary Verification Logic
```typescript
const handleVerify = async () => {
    // ... existing code ...
    
    if (useBinaryEncoding) {
        // Use binary encoding for verification
        const signatureBytes = DIDCrypto.hexToUint8Array(authenticationState.signature)
        const isValid = await verifyDIDBinary(didInfo.did.ID, signatureBytes, authenticationState.proof)
        setVerificationResult(isValid)
    } else {
        // Use string encoding for verification
        const isValid = await verifyDID(didInfo.did.ID, authenticationState.signature, authenticationState.proof)
        setVerificationResult(isValid)
    }
};
```

## 4. Comprehensive Usage Examples

### Basic Binary Operations
```typescript
import { DIDCrypto } from '@/app/api/did';

// Generate binary key pair
const keyPair = await DIDCrypto.generateKeyPairBinary();
console.log('Private key length:', keyPair.privateKey.length); // e.g., 32 bytes
console.log('Public key length:', keyPair.publicKey.length);   // e.g., 32 bytes

// Convert string message to binary
const message = "Hello, WebAssembly!";
const messageBytes = DIDCrypto.stringToUint8Array(message);
console.log('Message as binary:', messageBytes);

// Sign message with binary data
const signature = await DIDCrypto.signMessageBinary(keyPair.privateKey, messageBytes);
console.log('Signature length:', signature.length); // e.g., 64 bytes

// Verify signature
const isValid = await DIDCrypto.verifySignatureBinary(
    keyPair.publicKey, 
    messageBytes, 
    signature
);
console.log('Signature valid:', isValid); // true
```

### Data Conversion Examples
```typescript
// String ↔ Binary conversion
const text = "Hello, Binary World! 🌍";
const textAsBytes = DIDCrypto.stringToUint8Array(text);
const backToText = DIDCrypto.uint8ArrayToString(textAsBytes);
console.log('Round-trip successful:', text === backToText); // true

// Hex ↔ Binary conversion
const hexString = "48656c6c6f20576f726c64";
const hexAsBytes = DIDCrypto.hexToUint8Array(hexString);
const backToHex = DIDCrypto.uint8ArrayToHex(hexAsBytes);
console.log('Hex round-trip successful:', hexString === backToHex); // true

// Convert from hex string (like from storage) to binary for WASM
const storedPrivateKey = "a1b2c3d4e5f6..."; // from localStorage or database
const privateKeyBytes = DIDCrypto.hexToUint8Array(storedPrivateKey);
// Now use privateKeyBytes with binary WASM functions
```

### Performance Comparison
```typescript
// Binary encoding (optimized)
const startBinary = performance.now();
const keyPair = await DIDCrypto.generateKeyPairBinary();
const messageBytes = DIDCrypto.stringToUint8Array("test message");
const signature = await DIDCrypto.signMessageBinary(keyPair.privateKey, messageBytes);
const isValid = await DIDCrypto.verifySignatureBinary(keyPair.publicKey, messageBytes, signature);
const endBinary = performance.now();
console.log(`Binary encoding time: ${endBinary - startBinary}ms`);

// String encoding (legacy)
const startString = performance.now();
const { generateKeyPair, sign, verify } = await import('@/lib/crypto-wasm');
const keyPairStr = await generateKeyPair();
const signatureStr = await sign(keyPairStr.privateKey, "test message");
const isValidStr = await verify(keyPairStr.publicKey, "test message", signatureStr);
const endString = performance.now();
console.log(`String encoding time: ${endString - startString}ms`);
```

### Integration with DID Operations
```typescript
// Complete DID workflow with binary encoding
async function performDIDOperationsWithBinary() {
    // 1. Generate DID with binary keys
    const didResult = await createDID({ name: "Test User", dob: "1990-01-01" });
    
    // 2. Convert stored hex key to binary for operations
    const privateKeyBytes = DIDCrypto.hexToUint8Array(didResult.privateKey);
    
    // 3. Generate binary challenge
    const challengeBytes = await DIDCrypto.generateChallengeBinary(32);
    
    // 4. Authenticate with binary data
    const authResult = await authenticateDIDBinary(
        didResult.did.ID,
        privateKeyBytes,
        challengeBytes
    );
    
    // 5. Verify with binary signature
    const isValid = await verifyDIDBinary(
        didResult.did.ID,
        authResult.signature,
        authResult.proof
    );
    
    console.log('DID operations with binary encoding completed:', isValid);
    return { didResult, authResult, isValid };
}
```

## 5. Legacy Compatibility

### Backward Compatible Functions
```typescript
// Legacy functions still work with string parameters
export const generateKeyPair = async () => {
    const result = await cryptoWasm.generateKeyPair();
    return {
        privateKey: cryptoWasm.uint8ArrayToHex(result.privateKey),
        publicKey: cryptoWasm.uint8ArrayToHex(result.publicKey),
    };
};

export const sign = async (privateKey: string, message: string) => {
    const privateKeyBytes = cryptoWasm.hexToUint8Array(privateKey);
    const messageBytes = cryptoWasm.stringToUint8Array(message);
    const signature = await cryptoWasm.signMessage(privateKeyBytes, messageBytes);
    return cryptoWasm.uint8ArrayToHex(signature);
};

export const verify = async (publicKey: string, message: string, signature: string) => {
    const publicKeyBytes = cryptoWasm.hexToUint8Array(publicKey);
    const messageBytes = cryptoWasm.stringToUint8Array(message);
    const signatureBytes = cryptoWasm.hexToUint8Array(signature);
    return await cryptoWasm.verifySignature(publicKeyBytes, messageBytes, signatureBytes);
};
```

This implementation provides both high-performance binary operations and backward compatibility with existing string-based code.