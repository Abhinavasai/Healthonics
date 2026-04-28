package security

import (
	"bytes"
	"os"
	"testing"
)

func TestEnvelopeEncryptDecrypt_RoundTrip(t *testing.T) {
	_ = os.Setenv("PHI_ENCRYPTION_KEYS", "1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	_ = os.Setenv("PHI_ENCRYPTION_ACTIVE_VERSION", "1")
	t.Cleanup(func() {
		_ = os.Unsetenv("PHI_ENCRYPTION_KEYS")
		_ = os.Unsetenv("PHI_ENCRYPTION_ACTIVE_VERSION")
	})

	kr, err := LoadKeyringFromEnv()
	if err != nil {
		t.Fatalf("LoadKeyringFromEnv: %v", err)
	}

	plain := []byte("hello phi payload")
	env, err := kr.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	out, err := kr.Decrypt(*env)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("round trip mismatch")
	}
	if bytes.Equal(env.Ciphertext, plain) {
		t.Fatalf("ciphertext must not equal plaintext")
	}
}

func TestEnvelopeDecrypt_BackwardCompatibleKeyRotation(t *testing.T) {
	// Two-key keyring; active is v2, but we must still decrypt v1 envelopes.
	_ = os.Setenv("PHI_ENCRYPTION_KEYS", "1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=,2:BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=")
	_ = os.Setenv("PHI_ENCRYPTION_ACTIVE_VERSION", "2")
	t.Cleanup(func() {
		_ = os.Unsetenv("PHI_ENCRYPTION_KEYS")
		_ = os.Unsetenv("PHI_ENCRYPTION_ACTIVE_VERSION")
	})

	kr, err := LoadKeyringFromEnv()
	if err != nil {
		t.Fatalf("LoadKeyringFromEnv: %v", err)
	}

	// Force v1 encryption by overriding activeVersion.
	krV1 := *kr
	krV1.activeVersion = 1

	env, err := krV1.Encrypt([]byte("v1 secret"))
	if err != nil {
		t.Fatalf("Encrypt v1: %v", err)
	}
	if env.KeyVersion != 1 {
		t.Fatalf("expected v1 envelope")
	}

	// Decrypt with full keyring (active v2).
	out, err := kr.Decrypt(*env)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(out) != "v1 secret" {
		t.Fatalf("unexpected plaintext: %q", string(out))
	}

	// Encrypt with v2 and verify decrypt works.
	env2, err := kr.Encrypt([]byte("v2 secret"))
	if err != nil {
		t.Fatalf("Encrypt v2: %v", err)
	}
	if env2.KeyVersion != 2 {
		t.Fatalf("expected v2 envelope")
	}
	out2, err := kr.Decrypt(*env2)
	if err != nil {
		t.Fatalf("Decrypt v2: %v", err)
	}
	if string(out2) != "v2 secret" {
		t.Fatalf("unexpected plaintext: %q", string(out2))
	}
}
