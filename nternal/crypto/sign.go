package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
)

// Sha256Hash returns the SHA-256 hash of the input data.
func Sha256Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// SignData signs the given data using the provided private key.
// Returns a hex-encoded signature string.
func SignData(data []byte, privKey *ecdsa.PrivateKey) (string, error) {
	if privKey == nil {
		return "", fmt.Errorf("private key is nil")
	}
	hash := Sha256Hash(data)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return "", err
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature), nil
}

// VerifySignature verifies a signature against the original data and public key.
func VerifySignature(data []byte, signatureHex string, pubKey *ecdsa.PublicKey) (bool, error) {
	if pubKey == nil {
		return false, fmt.Errorf("public key is nil")
	}
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature hex: %w", err)
	}
	hash := Sha256Hash(data)
	r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
	s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])
	return ecdsa.Verify(pubKey, hash, r, s), nil
}

// GenerateKeyPair generates a new ECDSA key pair (P-256 curve).
// Returns the private key in PEM format.
func GenerateKeyPair() (string, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", err
	}
	privPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	}))
	return privPEM, nil
}

// LoadPrivateKeyFromPEM loads an ECDSA private key from a PEM string.
func LoadPrivateKeyFromPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return priv, nil
}

// PrivateKeyToHex returns the hex-encoded D parameter of the private key.
func PrivateKeyToHex(priv *ecdsa.PrivateKey) string {
	return hex.EncodeToString(priv.D.Bytes())
}

// PublicKeyToHex returns the hex-encoded representation of a public key.
func PublicKeyToHex(pub *ecdsa.PublicKey) string {
	return fmt.Sprintf("%x%x", pub.X.Bytes(), pub.Y.Bytes())
}
