package crypto

import (
	"testing"
)

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	plaintexts := [][]byte{
		[]byte("AKIAIOSFODNN7EXAMPLE"),
		[]byte("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		[]byte("test-key"),
		[]byte(""),
	}

	for _, plaintext := range plaintexts {
		encoded, err := Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt(%q) failed: %v", plaintext, err)
		}
		if encoded == "" {
			t.Fatal("expected non-empty encoded result")
		}

		decrypted, err := Decrypt(encoded)
		if err != nil {
			t.Fatalf("Decrypt failed: %v", err)
		}

		if string(decrypted) != string(plaintext) {
			t.Errorf("roundtrip mismatch: got %q, want %q", decrypted, plaintext)
		}
	}
}

func TestEncrypt_ProducesDifferentOutputs(t *testing.T) {
	plaintext := []byte("same-key-different-nonce")

	enc1, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("first Encrypt failed: %v", err)
	}
	enc2, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}

	if enc1 == enc2 {
		t.Error("expected different ciphertexts due to random nonce")
	}
}

func TestDecrypt_InvalidInput(t *testing.T) {
	_, err := Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	_, err = Decrypt("dG9vLXNob3J0")
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestDecrypt_CorruptedCiphertext(t *testing.T) {
	plaintext := []byte("sensitive-data")
	encoded, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	corrupted := encoded[:len(encoded)-1] + "X"
	_, err = Decrypt(corrupted)
	if err == nil {
		t.Error("expected error for corrupted ciphertext")
	}
}
