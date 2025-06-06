# API Operation Batching Implementation

This document describes the implementation of API operation batching in the Next.js DID application to improve performance and reduce network overhead.

## Overview

API operation batching combines multiple related API calls into single requests, reducing network latency and improving application performance. Two main batch operations have been implemented:

1. **DID Creation + Credential Issuance**: Combines DID creation and age credential issuance
2. **Authentication + Verification**: Combines DID authentication and verification

## Backend Implementation

### New Endpoints

#### Batch Create and Issue Credential
- **Endpoint**: `POST /api/did/batch/create-and-issue`
- **Purpose**: Creates a new DID and issues an age credential in a single operation
- **Request Body**:
  ```json
  {
    "name": "John Doe",
    "dob": "19900101"
  }
  ```
- **Response**:
  ```json
  {
    "did": { "ID": "did:example:123..." },
    "privateKey": "0x123...",
    "success": true,
    "message": "DID created and credential issued successfully"
  }
  ```

#### Batch Authenticate and Verify
- **Endpoint**: `POST /api/did/batch/auth-and-verify`
- **Purpose**: Authenticates a DID and verifies the authentication in a single operation
- **Request Body**:
  ```json
  {
    "didId": "did:example:123...",
    "privateKey": "0x123...",
    "challenge": "random_challenge"
  }
  ```
- **Response**:
  ```json
  {
    "proof": "0xabc...",
    "signature": "0xdef...",
    "verified": true,
    "success": true,
    "message": "Authentication and verification completed successfully"
  }
  ```

### Backend Code Changes

#### New Models (`backend/internal/did/models.go`)
```go
// Batch request models
type BatchCreateAndIssueRequest struct {
    Name string `json:"name" validate:"required,min=2,max=100,safe_string"`
    DOB  string `json:"dob" validate:"required,dob_format"`
}

type BatchAuthAndVerifyRequest struct {
    DIDID      string `json:"didId" validate:"required,did_format"`
    PrivateKey string `json:"privateKey" validate:"required,hex_string"`
    Challenge  string `json:"challenge" validate:"required,min=8,max=256,safe_string"`
}

// Batch response models
type BatchCreateAndIssueResponse struct {
    DID        interface{} `json:"did"`
    PrivateKey string      `json:"privateKey"`
    Success    bool        `json:"success"`
    Message    string      `json:"message,omitempty"`
}

type BatchAuthAndVerifyResponse struct {
    Proof     string `json:"proof"`
    Signature string `json:"signature"`
    Verified  bool   `json:"verified"`
    Success   bool   `json:"success"`
    Message   string `json:"message,omitempty"`
}
```

#### New Handlers (`backend/internal/did/handlers.go`)
- `BatchCreateAndIssue()`: Combines DID creation and credential issuance
- `BatchAuthAndVerify()`: Combines authentication and verification

#### Updated Routes (`backend/cmd/api/main.go`)
```go
// Batch DID operations
batchGroup := didGroup.Group("/batch")
batchGroup.Post("/create-and-issue", didHandlers.BatchCreateAndIssue)
batchGroup.Post("/auth-and-verify", didHandlers.BatchAuthAndVerify)
```

## Frontend Implementation

### Updated API Client (`client/src/app/api/did.ts`)

#### New Functions
```typescript
// Batch: Create DID and issue credential
export async function batchCreateDIDAndIssueCredential(
  formData: BatchCreateAndIssueRequest
): Promise<BatchCreateAndIssueResponse>

// Batch: Authenticate and verify DID
export async function batchAuthenticateAndVerifyDID(
  request: BatchAuthAndVerifyRequest
): Promise<BatchAuthAndVerifyResponse>

// Helper to determine when to use batching
export function shouldUseBatching(): boolean
```

### Updated UI Component (`client/src/components/DIDForm.tsx`)

#### New Features
- **Batching Toggle**: Users can enable/disable batch operations
- **Loading States**: Visual feedback during batch operations
- **Operation Indicators**: Shows which type of operation is being performed
- **Batch Success Messages**: Displays when batch operations complete successfully

#### UI Changes
- Toggle switch to enable/disable batching
- Updated button text to indicate batch operations
- Loading spinners with batch-specific messages
- Success indicators for completed batch operations

## Performance Benefits

### API Call Reduction
- **Individual Operations**: 3 API calls (Create DID → Authenticate → Verify)
- **Batch Operations**: 2 API calls (Create+Issue → Auth+Verify)
- **Reduction**: 33.3% fewer API calls

### Network Efficiency
- Reduced network round trips
- Lower latency for complete workflows
- Reduced bandwidth usage
- Better performance on slow connections

### Error Handling
- Atomic operations with rollback capability
- Better error reporting with context
- Reduced complexity in error handling

## Usage Examples

### Individual Operations (Traditional)
```typescript
// Step 1: Create DID
const didResponse = await createDID({ name: 'John', dob: '2000-01-01' });

// Step 2: Authenticate
const authResponse = await authenticateDID(
  didResponse.did.ID, 
  didResponse.privateKey, 
  'challenge'
);

// Step 3: Verify
const verified = await verifyDID(
  didResponse.did.ID, 
  authResponse.signature, 
  authResponse.proof
);
```

### Batch Operations (Optimized)
```typescript
// Step 1: Create DID and Issue Credential (Batch)
const batchCreateResponse = await batchCreateDIDAndIssueCredential({
  name: 'John',
  dob: '2000-01-01'
});

// Step 2: Authenticate and Verify (Batch)
const batchAuthResponse = await batchAuthenticateAndVerifyDID({
  didId: batchCreateResponse.did.ID,
  privateKey: batchCreateResponse.privateKey,
  challenge: 'challenge'
});
```

## When to Use Batch Operations

### Recommended Scenarios
- Production applications with high throughput
- Mobile applications with limited connectivity
- When operations are always performed together
- When you want to reduce API rate limiting impact

### Individual Operations Preferred
- Testing individual components
- Debugging specific operations
- When operations need to be performed at different times
- When you need to modify data between operations

## Error Handling

### Batch Operation Errors
Batch operations include comprehensive error handling:

```typescript
try {
  const response = await batchCreateDIDAndIssueCredential(data);
  if (!response.success) {
    throw new Error(response.message);
  }
  // Handle success
} catch (error) {
  // Handle error with context
  console.error('Batch operation failed:', error.message);
}
```

### Partial Failure Handling
- If DID creation succeeds but credential issuance fails, the response indicates partial success
- If authentication succeeds but verification fails, both proof and verification status are returned

## Testing

### Demo Script
A comprehensive demo script is available at `client/src/examples/batch-operations-demo.ts` that demonstrates:
- Individual vs batch operation performance
- Error handling scenarios
- Usage recommendations
- Performance comparisons

### Running the Demo
```typescript
import { 
  demonstrateIndividualOperations,
  demonstrateBatchOperations,
  performanceComparison 
} from './examples/batch-operations-demo';

// Run performance comparison
await performanceComparison();
```

## Configuration

### Environment Variables
No additional environment variables are required. The batching feature works with existing configuration.

### Feature Toggle
The `shouldUseBatching()` function can be modified to implement feature flags or A/B testing:

```typescript
export function shouldUseBatching(): boolean {
  // Can be controlled by feature flags, user preferences, etc.
  return process.env.NEXT_PUBLIC_ENABLE_BATCHING !== 'false';
}
```

## Future Enhancements

### Potential Improvements
1. **Adaptive Batching**: Automatically choose batch vs individual based on network conditions
2. **Batch Size Optimization**: Support for larger batch operations
3. **Caching**: Cache batch results for repeated operations
4. **Analytics**: Track performance improvements from batching
5. **Progressive Enhancement**: Fallback to individual operations if batch fails

### Monitoring
Consider adding metrics to track:
- Batch operation success rates
- Performance improvements
- Error rates comparison
- User adoption of batch features

## Security Considerations

- All batch operations maintain the same security validations as individual operations
- Input sanitization is applied to all batch request parameters
- Authentication and authorization requirements are preserved
- Batch operations include the same rate limiting as individual operations

## Conclusion

The API operation batching implementation provides significant performance improvements while maintaining backward compatibility with existing individual operations. The feature is designed to be transparent to users while providing better performance for production use cases.