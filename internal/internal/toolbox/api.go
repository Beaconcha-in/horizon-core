package toolbox

import (
	"encoding/json"
	"net/http"

	"horizon-core-engine/internal/crypto"
	"horizon-core-engine/internal/merkle"
	"horizon-core-engine/internal/network"
)

// ToolboxHandler هندلرهای API برای ابزارها
type ToolboxHandler struct {
	privateKey *crypto.ecdsa.PrivateKey
	publicKey  *crypto.ecdsa.PublicKey
}

// NewToolboxHandler ایجاد هندلر جدید
func NewToolboxHandler() *ToolboxHandler {
	return &ToolboxHandler{}
}

// GenerateKeyPairHandler تولید جفت کلید (API)
func (h *ToolboxHandler) GenerateKeyPairHandler(w http.ResponseWriter, r *http.Request) {
	privateKey, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"private_key": crypto.PrivateKeyToHex(privateKey),
		"public_key":  crypto.SHA256HashString(crypto.PrivateKeyToHex(privateKey)),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SignDataHandler امضای داده (API)
func (h *ToolboxHandler) SignDataHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.privateKey == nil {
		http.Error(w, "private key not set", http.StatusInternalServerError)
		return
	}

	signature, err := crypto.SignData(h.privateKey, []byte(req.Data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"signature": signature})
}

// EncryptHandler رمزنگاری AES (API)
func (h *ToolboxHandler) EncryptHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key       string `json:"key"`
		Plaintext string `json:"plaintext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key, err := hex.DecodeString(req.Key)
	if err != nil {
		http.Error(w, "invalid key format", http.StatusBadRequest)
		return
	}

	ciphertext, err := crypto.EncryptAES(key, req.Plaintext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ciphertext": ciphertext})
}

// DecryptHandler رمزگشایی AES (API)
func (h *ToolboxHandler) DecryptHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key        string `json:"key"`
		Ciphertext string `json:"ciphertext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key, err := hex.DecodeString(req.Key)
	if err != nil {
		http.Error(w, "invalid key format", http.StatusBadRequest)
		return
	}

	plaintext, err := crypto.DecryptAES(key, req.Ciphertext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"plaintext": plaintext})
}

// GenerateLicenseHandler تولید لایسنس (API)
func (h *ToolboxHandler) GenerateLicenseHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID string `json:"product_id"`
		UserID    string `json:"user_id"`
		Duration  int64  `json:"duration"` // ساعت
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.privateKey == nil {
		http.Error(w, "private key not set", http.StatusInternalServerError)
		return
	}

	generator := merkle.NewLicenseGenerator(h.privateKey)
	license, err := generator.GenerateLicense(req.ProductID, req.UserID, time.Duration(req.Duration)*time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(license)
}

// SubnetHandler محاسبه زیرنت (API)
func (h *ToolboxHandler) SubnetHandler(w http.ResponseWriter, r *http.Request) {
	cidr := r.URL.Query().Get("cidr")
	if cidr == "" {
		http.Error(w, "cidr parameter required", http.StatusBadRequest)
		return
	}

	info, err := network.CalculateSubnet(cidr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// MACLookupHandler جستجوی MAC (API)
func (h *ToolboxHandler) MACLookupHandler(w http.ResponseWriter, r *http.Request) {
	mac := r.URL.Query().Get("mac")
	if mac == "" {
		http.Error(w, "mac parameter required", http.StatusBadRequest)
		return
	}

	result := network.LookupOUI(mac)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"vendor": result})
}
