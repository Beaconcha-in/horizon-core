package crypto

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

// SHA256Hash محاسبه هش SHA-256
func SHA256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SHA256HashString محاسبه هش SHA-256 از رشته
func SHA256HashString(text string) string {
	return SHA256Hash([]byte(text))
}

// SHA512Hash محاسبه هش SHA-512
func SHA512Hash(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

// SHA512HashString محاسبه هش SHA-512 از رشته
func SHA512HashString(text string) string {
	return SHA512Hash([]byte(text))
}

// DoubleSHA256 محاسبه هش دوگانه SHA-256 (مانند بیت‌کوین)
func DoubleSHA256(data []byte) string {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return hex.EncodeToString(second[:])
}
