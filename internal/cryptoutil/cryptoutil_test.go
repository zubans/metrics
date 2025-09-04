package cryptoutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func generateRSAKeyPair(tb testing.TB) (*rsa.PrivateKey, *rsa.PublicKey) {
	b := tb
	b.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatalf("failed to generate rsa key: %v", err)
	}
	return priv, &priv.PublicKey
}

func writePrivateKeyPKCS1(tb testing.TB, dir string, priv *rsa.PrivateKey) string {
	b := tb
	b.Helper()
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	path := filepath.Join(dir, "priv_pkcs1.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		b.Fatalf("failed to write private key: %v", err)
	}
	return path
}

func writePrivateKeyPKCS8(tb testing.TB, dir string, priv *rsa.PrivateKey) string {
	b := tb
	b.Helper()
	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		b.Fatalf("failed to marshal pkcs8 private key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	path := filepath.Join(dir, "priv_pkcs8.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		b.Fatalf("failed to write private key: %v", err)
	}
	return path
}

func writePublicKeyPKIX(tb testing.TB, dir string, pub *rsa.PublicKey) string {
	b := tb
	b.Helper()
	spki, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		b.Fatalf("failed to marshal public key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: spki})
	path := filepath.Join(dir, "pub_pkix.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		b.Fatalf("failed to write public key: %v", err)
	}
	return path
}

func TestLoadKeys_PKCS1_PKIX(t *testing.T) {
	priv, pub := generateRSAKeyPair(t)
	dir := t.TempDir()
	pubPath := writePublicKeyPKIX(t, dir, pub)
	privPath := writePrivateKeyPKCS1(t, dir, priv)

	loadedPub, err := LoadPublicKey(pubPath)
	if err != nil {
		t.Fatalf("LoadPublicKey error: %v", err)
	}
	if loadedPub.N.Cmp(pub.N) != 0 || loadedPub.E != pub.E {
		t.Fatalf("loaded public key does not match")
	}

	loadedPriv, err := LoadPrivateKey(privPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey error: %v", err)
	}
	if loadedPriv.N.Cmp(priv.N) != 0 || loadedPriv.E != priv.E {
		t.Fatalf("loaded private key does not match")
	}
}

func TestLoadPrivateKey_PKCS8(t *testing.T) {
	priv, _ := generateRSAKeyPair(t)
	dir := t.TempDir()
	privPath := writePrivateKeyPKCS8(t, dir, priv)

	loadedPriv, err := LoadPrivateKey(privPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey(PKCS8) error: %v", err)
	}
	if loadedPriv.N.Cmp(priv.N) != 0 || loadedPriv.E != priv.E {
		t.Fatalf("loaded private key does not match")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	priv, pub := generateRSAKeyPair(t)
	plaintext := []byte("hello, metrics!")

	env, err := EncryptHybrid(pub, plaintext)
	if err != nil {
		t.Fatalf("EncryptHybrid error: %v", err)
	}

	got, err := DecryptHybrid(priv, env)
	if err != nil {
		t.Fatalf("DecryptHybrid error: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("roundtrip mismatch: got %q want %q", got, plaintext)
	}
}

func TestDecrypt_TamperedCiphertext_Fails(t *testing.T) {
	priv, pub := generateRSAKeyPair(t)
	env, err := EncryptHybrid(pub, []byte("payload"))
	if err != nil {
		t.Fatalf("EncryptHybrid error: %v", err)
	}
	// flip first byte of Data
	dataBytes, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	dataBytes[0] ^= 0xFF
	env.Data = base64.StdEncoding.EncodeToString(dataBytes)
	if _, err := DecryptHybrid(priv, env); err == nil {
		t.Fatalf("expected decryption error for tampered data, got nil")
	}
}

func TestDecrypt_WithWrongPrivateKey_Fails(t *testing.T) {
	_, pub1 := generateRSAKeyPair(t)
	priv2, _ := generateRSAKeyPair(t)
	env, err := EncryptHybrid(pub1, []byte("abc"))
	if err != nil {
		t.Fatalf("EncryptHybrid error: %v", err)
	}
	if _, err := DecryptHybrid(priv2, env); err == nil {
		t.Fatalf("expected error when decrypting with wrong private key")
	}
}

func TestLoadKeys_InvalidPEM(t *testing.T) {
	dir := t.TempDir()
	pubPath := filepath.Join(dir, "bad_pub.pem")
	privPath := filepath.Join(dir, "bad_priv.pem")
	if err := os.WriteFile(pubPath, []byte("not pem"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(privPath, []byte("not pem"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := LoadPublicKey(pubPath); err == nil {
		t.Fatalf("expected error for invalid public pem")
	}
	if _, err := LoadPrivateKey(privPath); err == nil {
		t.Fatalf("expected error for invalid private pem")
	}
}

func TestLoadPrivateKey_UnsupportedType(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ec_priv.pem")
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: []byte{1, 2, 3}})
	if err := os.WriteFile(p, pemBytes, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := LoadPrivateKey(p); err == nil {
		t.Fatalf("expected error for unsupported private key type")
	}
}
