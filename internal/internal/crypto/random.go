package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

// GenerateRandomBytes تولید بایت‌های تصادفی امن
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateRandomHex تولید رشته‌ی هگز تصادفی
func GenerateRandomHex(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomInt تولید عدد تصادفی امن در بازه [0, max)
func GenerateRandomInt(max int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

// GenerateUUID تولید UUID v4
func GenerateUUID() (string, error) {
	bytes, err := GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	// تنظیم نسخه 4 (UUID v4)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return hex.EncodeToString(bytes), nil
}
