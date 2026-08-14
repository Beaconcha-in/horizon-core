package license

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ============================================================
// ۱. ساختارهای داده
// ============================================================

type LicenseGuard struct {
	mu          sync.RWMutex
	licenseKey  string
	serverURL   string
	valid       bool
	lastCheck   time.Time
	checkTicker *time.Ticker
	stopChan    chan struct{}
}

type LicenseCheckResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Expiry  string `json:"expiry"`
}

// ============================================================
// ۲. تابع اصلی (ایجاد گارد)
// ============================================================

func NewLicenseGuard() *LicenseGuard {
	licenseKey := os.Getenv("HORIZON_LICENSE_KEY")
	serverURL := os.Getenv("HORIZON_LICENSE_SERVER")

	if licenseKey == "" || serverURL == "" {
		log.Fatal("❌ HORIZON_LICENSE_KEY and HORIZON_LICENSE_SERVER must be set")
	}

	return &LicenseGuard{
		licenseKey: licenseKey,
		serverURL:  serverURL,
		valid:      false,
		stopChan:   make(chan struct{}),
	}
}

// ============================================================
// ۳. بررسی لایسنس (ارسال به سرور)
// ============================================================

func (g *LicenseGuard) checkLicense() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	payload := map[string]string{"license": g.licenseKey}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(
		g.serverURL+"/api/v1/license/verify",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("خطا در اتصال به License Server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("پاسخ نامعتبر از سرور (کد %d)", resp.StatusCode)
	}

	var result LicenseCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("خطا در پردازش پاسخ: %v", err)
	}

	if !result.Valid {
		return fmt.Errorf("لایسنس نامعتبر است: %s", result.Message)
	}

	g.valid = true
	g.lastCheck = time.Now()
	log.Printf("✅ لایسنس معتبر است. انقضا: %s", result.Expiry)
	return nil
}

// ============================================================
// ۴. شروع گارد (اجرا در هنگام استارت)
// ============================================================

func (g *LicenseGuard) Start() error {
	log.Println("🔐 در حال بررسی لایسنس...")

	if err := g.checkLicense(); err != nil {
		return fmt.Errorf("❌ راه‌اندازی متوقف شد: %v", err)
	}

	// چک کردن هر ۲۴ ساعت
	g.checkTicker = time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-g.checkTicker.C:
				log.Println("🔄 بررسی مجدد لایسنس...")
				if err := g.checkLicense(); err != nil {
					log.Printf("❌ لایسنس نامعتبر شد: %v", err)
					g.valid = false
					// در اینجا می‌توانید سیستم را متوقف کنید
				}
			case <-g.stopChan:
				g.checkTicker.Stop()
				return
			}
		}
	}()

	return nil
}

// ============================================================
// ۵. توقف گارد
// ============================================================

func (g *LicenseGuard) Stop() {
	close(g.stopChan)
	log.Println("🛑 گارد لایسنس متوقف شد")
}

// ============================================================
// ۶. بررسی وضعیت (برای API)
// ============================================================

func (g *LicenseGuard) IsValid() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.valid
}

// ============================================================
// ۷. غیرفعال‌سازی اضطراری (از راه دور)
// ============================================================

func (g *LicenseGuard) EmergencyDisable() error {
	log.Println("🚨 درخواست غیرفعال‌سازی اضطراری...")

	// ارسال درخواست به License Server برای غیرفعال کردن این لایسنس
	payload := map[string]string{"license": g.licenseKey, "action": "disable"}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(
		g.serverURL+"/api/v1/license/emergency-disable",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("خطا در ارتباط با سرور: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("سرور درخواست را رد کرد (کد %d)", resp.StatusCode)
	}

	g.mu.Lock()
	g.valid = false
	g.mu.Unlock()

	log.Println("✅ لایسنس با موفقیت غیرفعال شد")
	return nil
}
