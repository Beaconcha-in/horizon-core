package merkle

import (
	"encoding/json"
	"time"

	"horizon-core-engine/internal/crypto"
)

// LicenseInfo اطلاعات لایسنس
type LicenseInfo struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
	Signature string    `json:"signature"`
	Proof     []string  `json:"proof"`
	RootHash  string    `json:"root_hash"`
}

// LicenseGenerator تولیدکننده لایسنس
type LicenseGenerator struct {
	privateKey interface{} // ECDSA private key
	tree       *Tree
}

// NewLicenseGenerator ایجاد تولیدکننده جدید
func NewLicenseGenerator(privateKey interface{}) *LicenseGenerator {
	return &LicenseGenerator{
		privateKey: privateKey,
	}
}

// GenerateLicense تولید لایسنس جدید
func (g *LicenseGenerator) GenerateLicense(productID, userID string, duration time.Duration) (*LicenseInfo, error) {
	// تولید داده‌های لایسنس
	data := map[string]interface{}{
		"product_id": productID,
		"user_id":    userID,
		"created_at": time.Now().UTC(),
		"expires_at": time.Now().UTC().Add(duration),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// ساخت درخت مرکل (برای سادگی، یک برگ)
	tree := NewTree([][]byte{jsonData})

	// امضای داده‌ها (با استفاده از کلید خصوصی)
	// توجه: در پیاده‌سازی واقعی، باید کلید خصوصی را از پیکربندی بخوانید
	signature, err := crypto.SignData(g.privateKey.(*crypto.ecdsa.PrivateKey), jsonData)
	if err != nil {
		return nil, err
	}

	license := &LicenseInfo{
		ID:        crypto.SHA256HashString(string(jsonData) + time.Now().String()),
		ProductID: productID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(duration),
		Active:    true,
		Signature: signature,
		RootHash:  tree.GetRootHash(),
		Proof:     tree.GenerateProof(0),
	}

	return license, nil
}

// VerifyLicense راستی‌آزمایی لایسنس
func VerifyLicense(license *LicenseInfo, publicKey interface{}) bool {
	// بررسی انقضا
	if time.Now().UTC().After(license.ExpiresAt) {
		return false
	}

	// بررسی فعال بودن
	if !license.Active {
		return false
	}

	// بازسازی داده‌ها
	data := map[string]interface{}{
		"product_id": license.ProductID,
		"user_id":    license.UserID,
		"created_at": license.CreatedAt,
		"expires_at": license.ExpiresAt,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false
	}

	// بررسی امضا
	valid := crypto.VerifySignature(publicKey.(*crypto.ecdsa.PublicKey), jsonData, license.Signature)
	if !valid {
		return false
	}

	// بررسی اثبات مرکل
	return VerifyProof(crypto.SHA256Hash(jsonData), license.Proof, license.RootHash)
}
