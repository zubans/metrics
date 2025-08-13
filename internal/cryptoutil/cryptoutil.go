package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

type Envelope struct {
	Key   string `json:"key"`
	Nonce string `json:"nonce"`
	Data  string `json:"data"`
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM public key")
	}
	if block.Type == "PUBLIC KEY" {
		pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		pub, ok := pubIfc.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not an RSA public key")
		}
		return pub, nil
	}
	if block.Type == "RSA PUBLIC KEY" {
		pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return pub, nil
	}
	return nil, errors.New("unsupported public key type")
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM private key")
	}
	if block.Type == "RSA PRIVATE KEY" {
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return priv, nil
	}
	if block.Type == "PRIVATE KEY" {
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		priv, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return priv, nil
	}
	return nil, errors.New("unsupported private key type")
}

func EncryptHybrid(pub *rsa.PublicKey, plaintext []byte) (*Envelope, error) {
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	blk, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	label := []byte(nil)
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, label)
	if err != nil {
		return nil, err
	}

	return &Envelope{
		Key:   base64.StdEncoding.EncodeToString(encryptedKey),
		Nonce: base64.StdEncoding.EncodeToString(nonce),
		Data:  base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

func DecryptHybrid(priv *rsa.PrivateKey, env *Envelope) ([]byte, error) {
	encryptedKey, err := base64.StdEncoding.DecodeString(env.Key)
	if err != nil {
		return nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(env.Nonce)
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		return nil, err
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encryptedKey, nil)
	if err != nil {
		return nil, err
	}
	blk, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
