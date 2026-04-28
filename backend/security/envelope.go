package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Envelope contains ciphertext plus the minimal metadata required to decrypt.
// WrappedKey is the per-payload DEK encrypted ("wrapped") by a KEK (versioned keyring).
type Envelope struct {
	KeyVersion   int
	WrappedKey   []byte
	WrappedNonce []byte
	DataNonce    []byte
	Ciphertext   []byte
}

type Keyring struct {
	activeVersion int
	keys          map[int][]byte
}

var (
	ErrKeyringNotConfigured = errors.New("encryption keyring not configured")
	ErrUnknownKeyVersion    = errors.New("unknown key version")
)

// LoadKeyringFromEnv loads a versioned keyring from environment variables.
//
// - PHI_ENCRYPTION_KEYS: comma-separated list "1:<b64>,2:<b64>" where each key is 32 bytes (AES-256).
// - PHI_ENCRYPTION_ACTIVE_VERSION: optional; if absent uses highest version in PHI_ENCRYPTION_KEYS.
func LoadKeyringFromEnv() (*Keyring, error) {
	raw := strings.TrimSpace(os.Getenv("PHI_ENCRYPTION_KEYS"))
	if raw == "" {
		return nil, ErrKeyringNotConfigured
	}
	keys := map[int][]byte{}
	var versions []int
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pair := strings.SplitN(part, ":", 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid PHI_ENCRYPTION_KEYS entry: %q", part)
		}
		v, err := strconv.Atoi(strings.TrimSpace(pair[0]))
		if err != nil || v <= 0 {
			return nil, fmt.Errorf("invalid key version: %q", pair[0])
		}
		keyRaw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(pair[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid key b64 for version %d", v)
		}
		if len(keyRaw) != 32 {
			return nil, fmt.Errorf("version %d key must be 32 bytes (got %d)", v, len(keyRaw))
		}
		keys[v] = keyRaw
		versions = append(versions, v)
	}
	if len(keys) == 0 {
		return nil, ErrKeyringNotConfigured
	}
	sort.Ints(versions)
	active := versions[len(versions)-1]
	if av := strings.TrimSpace(os.Getenv("PHI_ENCRYPTION_ACTIVE_VERSION")); av != "" {
		v, err := strconv.Atoi(av)
		if err != nil || v <= 0 {
			return nil, fmt.Errorf("invalid PHI_ENCRYPTION_ACTIVE_VERSION: %q", av)
		}
		if _, ok := keys[v]; !ok {
			return nil, fmt.Errorf("%w: %d", ErrUnknownKeyVersion, v)
		}
		active = v
	}
	return &Keyring{activeVersion: active, keys: keys}, nil
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

func aesgcm(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (kr *Keyring) Encrypt(plaintext []byte) (*Envelope, error) {
	if kr == nil || len(kr.keys) == 0 {
		return nil, ErrKeyringNotConfigured
	}
	kek, ok := kr.keys[kr.activeVersion]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrUnknownKeyVersion, kr.activeVersion)
	}

	dek, err := randomBytes(32)
	if err != nil {
		return nil, err
	}

	dataAead, err := aesgcm(dek)
	if err != nil {
		return nil, err
	}
	dataNonce, err := randomBytes(dataAead.NonceSize())
	if err != nil {
		return nil, err
	}
	ciphertext := dataAead.Seal(nil, dataNonce, plaintext, nil)

	wrapAead, err := aesgcm(kek)
	if err != nil {
		return nil, err
	}
	wrapNonce, err := randomBytes(wrapAead.NonceSize())
	if err != nil {
		return nil, err
	}
	wrappedKey := wrapAead.Seal(nil, wrapNonce, dek, nil)

	return &Envelope{
		KeyVersion:   kr.activeVersion,
		WrappedKey:   wrappedKey,
		WrappedNonce: wrapNonce,
		DataNonce:    dataNonce,
		Ciphertext:   ciphertext,
	}, nil
}

func (kr *Keyring) Decrypt(env Envelope) ([]byte, error) {
	if kr == nil || len(kr.keys) == 0 {
		return nil, ErrKeyringNotConfigured
	}
	kek, ok := kr.keys[env.KeyVersion]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrUnknownKeyVersion, env.KeyVersion)
	}
	wrapAead, err := aesgcm(kek)
	if err != nil {
		return nil, err
	}
	dek, err := wrapAead.Open(nil, env.WrappedNonce, env.WrappedKey, nil)
	if err != nil {
		return nil, err
	}
	dataAead, err := aesgcm(dek)
	if err != nil {
		return nil, err
	}
	return dataAead.Open(nil, env.DataNonce, env.Ciphertext, nil)
}

func B64Encode(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func B64Decode(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(s)
}
