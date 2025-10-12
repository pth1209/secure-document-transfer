/**
 * Client-side encryption utilities using Web Crypto API
 * Implements hybrid encryption for secure end-to-end file transfer
 * 
 * Encryption Flow:
 * 1. Generate a random AES-256 key for each file
 * 2. Encrypt file chunks with AES-GCM using unique IVs per chunk
 * 3. Encrypt the AES key with each recipient's RSA-2048 public key using RSA-OAEP
 * 4. Only recipients with their private keys can decrypt the AES key and access the file
 * 
 * This ensures:
 * - Files are encrypted client-side before upload
 * - Server never has access to unencrypted data
 * - Only intended recipients can decrypt files
 * - Perfect forward secrecy (unique keys per file)
 */

export interface EncryptedData {
  encryptedData: ArrayBuffer;
  iv: string; // Base64-encoded initialization vector
  encryptedKey: string; // Base64-encoded encrypted AES key
}

/**
 * Generate a random AES-256 key
 */
export async function generateAESKey(): Promise<CryptoKey> {
  return await window.crypto.subtle.generateKey(
    {
      name: 'AES-GCM',
      length: 256,
    },
    true, // extractable
    ['encrypt', 'decrypt']
  );
}

/**
 * Generate a random initialization vector (IV)
 */
export function generateIV(): Uint8Array {
  return window.crypto.getRandomValues(new Uint8Array(12)); // 96-bit IV for AES-GCM
}

/**
 * Encrypt data using AES-GCM
 */
export async function encryptData(
  data: ArrayBuffer,
  key: CryptoKey,
  iv: Uint8Array
): Promise<ArrayBuffer> {
  return await window.crypto.subtle.encrypt(
    {
      name: 'AES-GCM',
      iv: iv as BufferSource,
    },
    key,
    data
  );
}

/**
 * Convert ArrayBuffer to Base64 string
 */
export function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return window.btoa(binary);
}

/**
 * Convert Base64 string to ArrayBuffer
 */
export function base64ToArrayBuffer(base64: string): ArrayBuffer {
  const binaryString = window.atob(base64);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes.buffer;
}

/**
 * Export a CryptoKey to raw format
 */
export async function exportKey(key: CryptoKey): Promise<ArrayBuffer> {
  return await window.crypto.subtle.exportKey('raw', key);
}

/**
 * Import a raw key as a CryptoKey
 */
export async function importKey(keyData: ArrayBuffer): Promise<CryptoKey> {
  return await window.crypto.subtle.importKey(
    'raw',
    keyData,
    {
      name: 'AES-GCM',
      length: 256,
    },
    true,
    ['encrypt', 'decrypt']
  );
}

/**
 * Generate RSA key pair for recipient key encryption
 * Note: In a real implementation, recipients would have their public keys stored
 * For this demo, we'll use a simplified approach with password-based encryption
 */
export async function generateRSAKeyPair(): Promise<CryptoKeyPair> {
  return await window.crypto.subtle.generateKey(
    {
      name: 'RSA-OAEP',
      modulusLength: 2048,
      publicExponent: new Uint8Array([1, 0, 1]),
      hash: 'SHA-256',
    },
    true,
    ['encrypt', 'decrypt']
  );
}

/**
 * Import an RSA public key from base64-encoded PEM format
 * The backend stores PEM keys as base64, so we need to decode that first
 */
export async function importRSAPublicKey(base64EncodedPem: string): Promise<CryptoKey> {
  // First, decode the base64 to get the actual PEM string
  const pemString = window.atob(base64EncodedPem);
  
  // Remove PEM header/footer and whitespace to get just the key data
  const pemContents = pemString
    .replace('-----BEGIN PUBLIC KEY-----', '')
    .replace('-----END PUBLIC KEY-----', '')
    .replace(/\s/g, '');
  
  // Convert base64 to binary
  const binaryKey = base64ToArrayBuffer(pemContents);
  
  // Import as RSA-OAEP key
  return await window.crypto.subtle.importKey(
    'spki',
    binaryKey,
    {
      name: 'RSA-OAEP',
      hash: 'SHA-256',
    },
    true,
    ['encrypt']
  );
}

/**
 * Encrypt AES key with RSA public key using RSA-OAEP
 */
export async function encryptKeyForRecipient(
  aesKey: CryptoKey,
  recipientPublicKeyPEM: string
): Promise<string> {
  // Export the AES key to raw format
  const keyData = await exportKey(aesKey);
  
  // Import the recipient's RSA public key
  const publicKey = await importRSAPublicKey(recipientPublicKeyPEM);
  
  // Encrypt the AES key with the recipient's RSA public key
  const encryptedKey = await window.crypto.subtle.encrypt(
    {
      name: 'RSA-OAEP',
    },
    publicKey,
    keyData
  );
  
  // Return as base64
  return arrayBufferToBase64(encryptedKey);
}

/**
 * Encrypt a file chunk
 */
export async function encryptChunk(
  chunkData: Blob,
  aesKey: CryptoKey,
  iv: Uint8Array
): Promise<ArrayBuffer> {
  // Convert Blob to ArrayBuffer
  const arrayBuffer = await chunkData.arrayBuffer();
  
  // Encrypt the data
  return await encryptData(arrayBuffer, aesKey, iv);
}

/**
 * Get MIME type from file
 */
export function getMimeType(file: File): string {
  return file.type || 'application/octet-stream';
}

