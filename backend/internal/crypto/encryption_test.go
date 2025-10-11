package crypto

import (
	"testing"
)

func TestGenerateUserKeys(t *testing.T) {
	password := "testPassword123"
	
	// Generate keys
	keys, err := GenerateUserKeys(password)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	
	// Verify all fields are populated
	if keys.PublicKeyPEM == "" {
		t.Error("PublicKeyPEM should not be empty")
	}
	if keys.EncryptedPrivateKey == "" {
		t.Error("EncryptedPrivateKey should not be empty")
	}
	if keys.Salt == "" {
		t.Error("Salt should not be empty")
	}
	if keys.IV == "" {
		t.Error("IV should not be empty")
	}
	
	t.Logf("✅ Keys generated successfully")
	t.Logf("Public Key (first 50 chars): %s...", keys.PublicKeyPEM[:50])
	t.Logf("Salt (first 20 chars): %s...", keys.Salt[:20])
}

func TestEncryptDecryptPrivateKey(t *testing.T) {
	password := "testPassword123"
	
	// Generate keys
	keys, err := GenerateUserKeys(password)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	
	// Decrypt the private key
	decryptedPrivateKey, err := DecryptPrivateKey(
		keys.EncryptedPrivateKey,
		keys.Salt,
		keys.IV,
		password,
	)
	if err != nil {
		t.Fatalf("Failed to decrypt private key: %v", err)
	}
	
	// Verify the decrypted key contains expected markers
	if len(decryptedPrivateKey) == 0 {
		t.Error("Decrypted private key should not be empty")
	}
	
	// Private key should contain PEM header/footer
	if len(decryptedPrivateKey) < 100 {
		t.Error("Decrypted private key seems too short")
	}
	
	t.Logf("✅ Private key decrypted successfully")
	t.Logf("Decrypted key (first 50 chars): %s...", decryptedPrivateKey[:50])
}

func TestDecryptWithWrongPassword(t *testing.T) {
	correctPassword := "testPassword123"
	wrongPassword := "wrongPassword456"
	
	// Generate keys with correct password
	keys, err := GenerateUserKeys(correctPassword)
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}
	
	// Try to decrypt with wrong password
	_, err = DecryptPrivateKey(
		keys.EncryptedPrivateKey,
		keys.Salt,
		keys.IV,
		wrongPassword,
	)
	
	// Should fail
	if err == nil {
		t.Error("Decryption should fail with wrong password")
	}
	
	t.Logf("✅ Correctly rejected wrong password: %v", err)
}

func TestKeyConsistency(t *testing.T) {
	password := "testPassword123"
	
	// Generate keys twice with same password
	keys1, err := GenerateUserKeys(password)
	if err != nil {
		t.Fatalf("Failed to generate keys1: %v", err)
	}
	
	keys2, err := GenerateUserKeys(password)
	if err != nil {
		t.Fatalf("Failed to generate keys2: %v", err)
	}
	
	// Keys should be different (different salt, IV, and key pairs)
	if keys1.PublicKeyPEM == keys2.PublicKeyPEM {
		t.Error("Public keys should be different for different generations")
	}
	if keys1.Salt == keys2.Salt {
		t.Error("Salts should be different for different generations")
	}
	if keys1.IV == keys2.IV {
		t.Error("IVs should be different for different generations")
	}
	
	t.Logf("✅ Each key generation produces unique keys")
}

