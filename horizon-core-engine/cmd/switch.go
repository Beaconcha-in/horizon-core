<!DOCTYPE html>
<html dir="rtl" lang="fa">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>اضافه کردن Switch به مخزن + GitHub Actions</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box}
        body{background:#060b18;color:#eef2ff;font-family:'Segoe UI',sans-serif;padding:2rem;display:flex;justify-content:center;align-items:center;min-height:100vh}
        .container{max-width:1100px;width:100%;background:rgba(12,18,34,0.88);backdrop-filter:blur(14px);border-radius:2rem;padding:2rem;border:1px solid rgba(79,158,255,0.15);box-shadow:0 25px 60px rgba(0,0,0,0.8)}
        h1{font-size:1.8rem;font-weight:900;background:linear-gradient(135deg,#fff,#4f9eff,#c084fc);-webkit-background-clip:text;background-clip:text;color:transparent;margin-bottom:0.3rem;display:flex;align-items:center;gap:0.6rem}
        .sub{color:#8fa2dc;font-size:0.9rem;margin-bottom:1.5rem;border-right:3px solid #4f9eff;padding-right:1rem}
        .file-section{background:#0a0f1e;border-radius:1.2rem;border:1px solid #1a2240;margin-bottom:1.5rem;overflow:hidden}
        .file-header{display:flex;justify-content:space-between;align-items:center;padding:0.6rem 1.2rem;background:#111a2e;border-bottom:1px solid #1a2240;flex-wrap:wrap;gap:0.5rem}
        .file-header .path{font-size:0.8rem;color:#8fa2dc;font-family:monospace}
        .btn-copy{background:#4f9eff;border:none;padding:0.3rem 1.2rem;border-radius:60px;font-weight:700;font-size:0.75rem;color:#0a0f1e;cursor:pointer;transition:0.3s;display:inline-flex;align-items:center;gap:0.4rem}
        .btn-copy:hover{background:#3a7bd5;transform:scale(1.03)}
        .btn-copy.copied{background:#4ade80}
        .code-body{padding:1rem;overflow-x:auto;max-height:400px;overflow-y:auto}
        .code-body pre{margin:0;font-family:'JetBrains Mono','Fira Code',monospace;font-size:0.75rem;line-height:1.7;color:#c8d6e5;white-space:pre;direction:ltr;text-align:left}
        .footer{margin-top:1.5rem;text-align:center;color:#475569;font-size:0.7rem;border-top:1px solid #1a2240;padding-top:1rem}
        .toast{position:fixed;bottom:30px;right:30px;background:rgba(20,26,44,0.95);backdrop-filter:blur(12px);border:1px solid #4ade80;padding:0.9rem 1.8rem;border-radius:1rem;color:#eef2ff;font-size:0.9rem;box-shadow:0 12px 40px rgba(0,0,0,0.7);z-index:9999;display:none;animation:slideUp 0.35s ease-out}
        .toast.error{border-color:#f87171}
        @keyframes slideUp{0%{opacity:0;transform:translateY(30px)}100%{opacity:1;transform:translateY(0)}}
        ::-webkit-scrollbar{width:5px;height:5px}
        ::-webkit-scrollbar-track{background:#0f172a}
        ::-webkit-scrollbar-thumb{background:#4f9eff;border-radius:4px}
        .emoji{font-size:1.2rem}
        .note{background:#111a2e;padding:0.8rem 1.2rem;border-radius:0.8rem;border-right:3px solid #facc15;margin-bottom:1.5rem;font-size:0.85rem;color:#c8d6e5}
        .note strong{color:#facc15}
    </style>
</head>
<body>

<div class="container">
    <h1>🔀 Horizon Switch + GitHub Actions</h1>
    <div class="sub">اضافه کردن سوئیچ به مخزن و استقرار خودکار با GitHub Actions</div>

    <div class="note">
        <strong>📌 راهنما:</strong> هر بخش شامل یک فایل با مسیر مشخص است. 
        فایل‌ها را به‌ترتیب در مسیرهای گفته‌شده در مخزن <code style="color:#4ade80;">horizon-core-engine</code> ایجاد کنید.
    </div>

    <!-- ============================================ -->
    <!-- فایل ۱: switch.go -->
    <!-- ============================================ -->
    <div class="file-section">
        <div class="file-header">
            <span class="path">📄 cmd/switch/switch.go</span>
            <button class="btn-copy" onclick="copyCode('code1')"><i class="fas fa-copy"></i> کپی</button>
        </div>
        <div class="code-body">
            <pre id="code1">package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type SwitchConfig struct {
	ListenPort       string            `json:"listen_port"`
	CoreEngineURL    string            `json:"core_engine_url"`
	LicenseServerURL string            `json:"license_server_url"`
	APIKey           string            `json:"api_key"`
	AdminToken       string            `json:"admin_token"`
	TargetServers    map[string]string `json:"target_servers"`
}

type Switch struct {
	config     *SwitchConfig
	httpClient *http.Client
	router     *mux.Router
	server     *http.Server
}

func SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func AESEncrypt(key, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

func AESDecrypt(key []byte, ciphertextHex string) ([]byte, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

type CoreEngineClient struct {
	baseURL string
	apiKey  string
}

func NewCoreEngineClient(baseURL, apiKey string) *CoreEngineClient {
	return &CoreEngineClient{baseURL: baseURL, apiKey: apiKey}
}

func (c *CoreEngineClient) GetBranches() ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/branches", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *CoreEngineClient) GetTransactions(limit int) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/transactions?limit=%d", c.baseURL, limit), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type LicenseClient struct {
	baseURL string
}

func NewLicenseClient(baseURL string) *LicenseClient {
	return &LicenseClient{baseURL: baseURL}
}

func (l *LicenseClient) VerifyLicense(licenseKey string) (bool, error) {
	body := map[string]string{"license": licenseKey}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", l.baseURL+"/api/v1/license/verify", bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	valid, ok := result["valid"].(bool)
	return ok && valid, nil
}

func (sw *Switch) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "online",
		"service": "Horizon Switch",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (sw *Switch) proxyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetName := vars["target"]
	targetURL, exists := sw.config.TargetServers[targetName]
	if !exists {
		http.Error(w, fmt.Sprintf("هدف '%s' یافت نشد", targetName), http.StatusNotFound)
		return
	}
	proxyReq, err := http.NewRequest(r.Method, targetURL+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxyReq.Header = r.Header.Clone()
	resp, err := sw.httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (sw *Switch) licenseCheckHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LicenseKey string `json:"license_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	client := NewLicenseClient(sw.config.LicenseServerURL)
	valid, err := client.VerifyLicense(req.LicenseKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid": valid,
		"time":  time.Now().Format(time.RFC3339),
	})
}

func (sw *Switch) encryptHandler(w http.ResponseWriter, r *http.Request) {
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
	ciphertext, err := AESEncrypt(key, []byte(req.Plaintext))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ciphertext": ciphertext})
}

func (sw *Switch) decryptHandler(w http.ResponseWriter, r *http.Request) {
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
	plaintext, err := AESDecrypt(key, req.Ciphertext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"plaintext": string(plaintext)})
}

func NewSwitch(config *SwitchConfig) *Switch {
	sw := &Switch{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		router: mux.NewRouter(),
	}
	sw.setupRoutes()
	return sw
}

func (sw *Switch) setupRoutes() {
	sw.router.HandleFunc("/health", sw.healthCheckHandler).Methods("GET")
	sw.router.HandleFunc("/api/v1/license/check", sw.licenseCheckHandler).Methods("POST")
	sw.router.HandleFunc("/api/v1/toolbox/encrypt", sw.encryptHandler).Methods("POST")
	sw.router.HandleFunc("/api/v1/toolbox/decrypt", sw.decryptHandler).Methods("POST")
	sw.router.HandleFunc("/proxy/{target}/{path:.*}", sw.proxyHandler).Methods("GET", "POST", "PUT", "DELETE")
	sw.router.HandleFunc("/proxy/{target}", sw.proxyHandler).Methods("GET", "POST", "PUT", "DELETE")
	sw.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "Horizon Switch",
			"version": "1.0.0",
			"endpoints": []string{
				"/health",
				"/api/v1/license/check",
				"/api/v1/toolbox/encrypt",
				"/api/v1/toolbox/decrypt",
				"/proxy/{target}",
				"/proxy/{target}/{path:.*}",
			},
		})
	})
}

func (sw *Switch) Start() error {
	sw.server = &http.Server{
		Addr:         ":" + sw.config.ListenPort,
		Handler:      sw.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("🚀 سوئیچ Horizon روی پورت %s راه‌اندازی شد", sw.config.ListenPort)
	log.Printf("🔗 متصل به Core Engine: %s", sw.config.CoreEngineURL)
	log.Printf("🔑 متصل به License Server: %s", sw.config.LicenseServerURL)
	return sw.server.ListenAndServe()
}

func (sw *Switch) Shutdown(ctx context.Context) error {
	log.Println("⏳ در حال توقف سوئیچ...")
	return sw.server.Shutdown(ctx)
}

func main() {
	listenPort := os.Getenv("SWITCH_PORT")
	if listenPort == "" {
		listenPort = "8080"
	}
	coreEngineURL := os.Getenv("CORE_ENGINE_URL")
	if coreEngineURL == "" {
		coreEngineURL = "http://localhost:3000"
	}
	licenseServerURL := os.Getenv("LICENSE_SERVER_URL")
	if licenseServerURL == "" {
		licenseServerURL = "http://localhost:4000"
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = "test-admin-token"
	}
	targetServers := map[string]string{
		"core":      coreEngineURL,
		"license":   licenseServerURL,
		"dashboard": "https://beaconchain-horizon.github.io/horizon-core-engine",
	}
	config := &SwitchConfig{
		ListenPort:       listenPort,
		CoreEngineURL:    coreEngineURL,
		LicenseServerURL: licenseServerURL,
		APIKey:           apiKey,
		AdminToken:       adminToken,
		TargetServers:    targetServers,
	}
	sw := NewSwitch(config)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := sw.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ خطا در اجرای سوئیچ: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("⏳ دریافت سیگنال خاتمه...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sw.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ خطا در توقف سوئیچ: %v", err)
	}
	log.Println("✅ سوئیچ متوقف شد")
}</pre>
        </div>
    </div>

    <!-- ============================================ -->
    <!-- فایل ۲: Dockerfile -->
    <!-- ============================================ -->
    <div class="file-section">
        <div class="file-header">
            <span class="path">🐳 cmd/switch/Dockerfile</span>
            <button class="btn-copy" onclick="copyCode('code2')"><i class="fas fa-copy"></i> کپی</button>
        </div>
        <div class="code-body">
            <pre id="code2">FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /switch ./cmd/switch

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /switch .
EXPOSE 8080
CMD ["./switch"]</pre>
        </div>
    </div>

    <!-- ============================================ -->
    <!-- فایل ۳: GitHub Actions -->
    <!-- ============================================ -->
    <div class="file-section">
        <div class="file-header">
            <span class="path">⚙️ .github/workflows/deploy-switch.yml</span>
            <button class="btn-copy" onclick="copyCode('code3')"><i class="fas fa-copy"></i> کپی</button>
        </div>
        <div class="code-body">
            <pre id="code3">name: Deploy Horizon Switch
on:
  push:
    branches: [ main ]
    paths:
      - 'cmd/switch/**'
      - '.github/workflows/deploy-switch.yml'

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build switch binary
        run: go build -o switch ./cmd/switch

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_TOKEN }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./cmd/switch/Dockerfile
          push: true
          tags: ${{ secrets.DOCKER_USERNAME }}/horizon-switch:latest

      - name: Deploy to server via SSH
        uses: appleboy/ssh-action@v1.0.0
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            docker pull ${{ secrets.DOCKER_USERNAME }}/horizon-switch:latest
            docker stop horizon-switch || true
            docker rm horizon-switch || true
            docker run -d --name horizon-switch -p 8080:8080 \
              -e CORE_ENGINE_URL=${{ secrets.CORE_ENGINE_URL }} \
              -e LICENSE_SERVER_URL=${{ secrets.LICENSE_SERVER_URL }} \
              -e API_KEY=${{ secrets.API_KEY }} \
              -e ADMIN_TOKEN=${{ secrets.ADMIN_TOKEN }} \
              ${{ secrets.DOCKER_USERNAME }}/horizon-switch:latest</pre>
        </div>
    </div>

    <!-- ============================================ -->
    <!-- جمع‌بندی و دستورات نهایی -->
    <!-- ============================================ -->
    <div class="note" style="border-right-color:#4ade80;">
        <strong>✅ بعد از کپی کردن فایل‌ها:</strong><br>
        ۱. هر فایل را در مسیر مشخص‌شده داخل مخزن <code style="color:#4ade80;">horizon-core-engine</code> ایجاد کنید.<br>
        ۲. متغیرهای مخفی (Secrets) را در تنظیمات مخزن GitHub اضافه کنید:<br>
        <code style="color:#4ade80;display:block;margin-top:0.3rem;background:#0f172a;padding:0.5rem;border-radius:0.5rem;">
        DOCKER_USERNAME, DOCKER_TOKEN, SERVER_HOST, SERVER_USER, SSH_PRIVATE_KEY, CORE_ENGINE_URL, LICENSE_SERVER_URL, API_KEY, ADMIN_TOKEN
        </code>
        ۳. تغییرات را <code style="color:#facc15;">git add</code> و <code style="color:#facc15;">git commit</code> و <code style="color:#facc15;">git push</code> کنید.<br>
        ۴. GitHub Actions خودکار اجرا می‌شود و سوئیچ روی سرور شما مستقر می‌شود.
    </div>

    <div class="footer">
        <span style="color:#4f9eff;">Horizon Switch</span> · آماده برای استقرار خودکار با GitHub Actions
    </div>
</div>

<div id="toast" class="toast">✅ کپی شد</div>

<script>
    function copyCode(id) {
        const el = document.getElementById(id);
        const code = el.textContent;
        const btn = event.target.closest('.btn-copy');

        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(code)
                .then(() => showToast('✅ کپی شد', false))
                .catch(() => fallbackCopy(code));
        } else {
            fallbackCopy(code);
        }
    }

    function fallbackCopy(text) {
        const ta = document.createElement('textarea');
        ta.value = text;
        ta.style.position = 'fixed';
        ta.style.opacity = '0';
        ta.style.left = '-9999px';
        document.body.appendChild(ta);
        ta.select();
        try {
            document.execCommand('copy');
            showToast('✅ کپی شد', false);
        } catch (e) {
            showToast('❌ کپی نشد', true);
        }
        document.body.removeChild(ta);
    }

    function showToast(msg, isError) {
        const t = document.getElementById('toast');
        t.textContent = msg;
        t.className = 'toast' + (isError ? ' error' : '');
        t.style.display = 'block';
        clearTimeout(t._hide);
        t._hide = setTimeout(() => t.style.display = 'none', 3000);
    }
</script>
</body>
</html>
