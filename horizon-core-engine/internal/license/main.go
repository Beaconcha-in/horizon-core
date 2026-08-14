package main

import (
	"log"
	"horizon-core-engine/internal/license"
)

func main() {
	// راه‌اندازی License Guard
	guard := license.NewLicenseGuard()
	if err := guard.Start(); err != nil {
		log.Fatalf("❌ %v", err)
	}
	defer guard.Stop()

	// بقیه کدهای Core Engine...
	// اگر guard.IsValid() false باشد، سیستم نباید تراکنش‌ها را پردازش کند.
}
