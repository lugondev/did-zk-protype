import nacl from 'tweetnacl'
import naclUtil from 'tweetnacl-util'
import { ed25519 } from '@noble/curves/ed25519';

export interface DID {
	ID: string;
	PublicKey: any;
	Document: {
		Context: string[];
		ID: string;
		Controller: string;
		Authentication: Array<{
			ID: string;
			Type: string;
			Controller: string;
			PublicKeyJwk: Record<string, any>;
		}>;
		Credentials: Array<{
			Context: string[];
			ID: string;
			Type: string[];
			Issuer: string;
			Subject: string;
			Claims: Record<string, any>;
			Proof: {
				Type: string;
				Created: string;
				VerificationMethod: string;
				ProofValue: any;
			};
		}>;
	};
}

interface CreateDIDRequest {
	name: string;
	dob: string;  // Changed to string to match backend expectation
}

interface CreateDIDResponse {
	did: DID;
	privateKey: string;
}

interface AuthResponse {
	proof: string;
	signature: string;
}

interface VerifyResponse {
	verified: boolean;
}

const convertDateToInt = (dateStr: string): number => {
	// Validate date format
	if (!/^\d{4}-\d{2}-\d{2}$/.test(dateStr)) {
		throw new Error('Invalid date format. Expected YYYY-MM-DD');
	}

	const date = new Date(dateStr);
	if (isNaN(date.getTime())) {
		throw new Error('Invalid date');
	}

	// Convert YYYY-MM-DD to YYYYMMDD
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const day = String(date.getDate()).padStart(2, '0');

	const dateInt = parseInt(`${year}${month}${day}`);

	// Validate range (19000101 to 20991231)
	if (dateInt < 19000101 || dateInt > 20991231) {
		throw new Error('Date must be between 1900-01-01 and 2099-12-31');
	}

	return dateInt;
};

export const createDID = async (data: {
	name: string;
	dob: string;
}): Promise<CreateDIDResponse> => {
	const formattedData: CreateDIDRequest = {
		name: data.name,
		dob: convertDateToInt(data.dob).toString(),  // Convert to string
	};

	const response = await fetch('http://localhost:8080/api/did/create', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(formattedData),
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Failed to create DID');
	}

	return response.json();
};

function hexToUint8Array(hexString: string): Uint8Array {
	return new Uint8Array(
		hexString.match(/.{1,2}/g)?.map(byte => parseInt(byte, 16)) || []
	);
}

function hexToBytes(hexString: string) {
	// Loại bỏ tiền tố 0x nếu có
	if (hexString.startsWith('0x')) {
		hexString = hexString.slice(2);
	}

	// Đảm bảo chuỗi hex có độ dài chẵn
	if (hexString.length % 2 !== 0) {
		hexString = '0' + hexString;
	}

	const bytes = new Uint8Array(hexString.length / 2);
	for (let i = 0; i < hexString.length; i += 2) {
		bytes[i / 2] = parseInt(hexString.substr(i, 2), 16);
	}

	return bytes;
}

async function signMessage(privateKeyHex: string, message: string) {
	// Chuyển đổi private key từ hex
	const privateKey = hexToBytes(privateKeyHex);
	console.log('privateKey', privateKey);

	const messageBytes = new TextEncoder().encode(message);
	console.log('messageBytes', messageBytes);

	// Hoặc nếu BN254 (tùy thuộc vào hỗ trợ cụ thể của thư viện)
	// const sig = bn254.sign(messageBytes, privateKey);
	const sig = await ed25519.sign(messageBytes, privateKey);

	console.log('sig', sig);

	// convert Uint8Array to hex string
	return Buffer.from(sig).toString('hex');
}

async function signChallenge(privateKeyHex: string, message: string): Promise<string> {
	// Convert private key from hex and use first 32 bytes as seed
	const privateKeyBytes = hexToUint8Array(privateKeyHex);
	const keyPair = nacl.sign.keyPair.fromSeed(privateKeyBytes.slice(0, 32));

	// Convert message to bytes
	const messageBytes = naclUtil.decodeUTF8(message);

	// Sign message using detached signature
	const signature = nacl.sign.detached(messageBytes, keyPair.secretKey);

	// Return signature as base64 string
	return naclUtil.encodeBase64(signature);
}

export const authenticateDID = async (
	didId: string,
	privateKey: string,
	challenge: string
): Promise<AuthResponse> => {
	// Sign the challenge first
	// const signature = await signMessage(privateKey, challenge);

	const response = await fetch('http://localhost:8080/api/did/authenticate', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			didId,
			privateKey,
			challenge,
			// signature,
		}),
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Failed to authenticate DID');
	}

	const data = await response.json();
	return {
		proof: data.proof,
		signature: data.signature, // Use the signature we generated client-side
	};
};

export const verifyDID = async (
	didId: string,
	signature: string,
	proof: string
): Promise<boolean> => {
	const response = await fetch('http://localhost:8080/api/did/verify', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			didId,
			signature,
			proof,
		}),
	});

	if (!response.ok) {
		const error = await response.text();
		throw new Error(error || 'Failed to verify DID');
	}

	const result: VerifyResponse = await response.json();
	return result.verified;
};
