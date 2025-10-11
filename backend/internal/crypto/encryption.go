package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// RSAKeySize is the RSA key size in bits
	RSAKeySize = 2048
	// PBKDF2Iterations for key derivation
	PBKDF2Iterations = 100000
	// SaltSize in bytes
	SaltSize = 32
	// AESKeySize in bytes (256 bits)
	AESKeySize = 32
)

// GeneratedKeys represents the keys generated for a user
type GeneratedKeys struct {
	PublicKeyPEM        string
	EncryptedPrivateKey string
	Salt                string
	IV                  string
}

// GenerateUserKeys generates an RSA key pair and encrypts the private key with a password-derived key
func GenerateUserKeys(password string) (*GeneratedKeys, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Convert private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Convert public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Generate random salt
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key from password using PBKDF2
	encryptionKey := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, AESKeySize, sha256.New)

	// Encrypt the private key with the derived key
	encryptedPrivateKey, iv, err := encryptAES(privateKeyPEM, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Encode everything to base64 for storage
	return &GeneratedKeys{
		PublicKeyPEM:        base64.StdEncoding.EncodeToString(publicKeyPEM),
		EncryptedPrivateKey: base64.StdEncoding.EncodeToString(encryptedPrivateKey),
		Salt:                base64.StdEncoding.EncodeToString(salt),
		IV:                  base64.StdEncoding.EncodeToString(iv),
	}, nil
}

// encryptAES encrypts data using AES-256-GCM
func encryptAES(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce (IV)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// DecryptPrivateKey decrypts the private key using the password
// This function is provided for reference/testing but should be called on the client-side
func DecryptPrivateKey(encryptedPrivateKeyBase64, saltBase64, ivBase64, password string) (string, error) {
	// Decode from base64
	encryptedPrivateKey, err := base64.StdEncoding.DecodeString(encryptedPrivateKeyBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted private key: %w", err)
	}

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	// Derive the same encryption key from password and salt
	encryptionKey := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, AESKeySize, sha256.New)

	// Decrypt the private key
	privateKeyPEM, err := decryptAES(encryptedPrivateKey, encryptionKey, iv)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return string(privateKeyPEM), nil
}

// decryptAES decrypts data using AES-256-GCM
func decryptAES(ciphertext []byte, key []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

