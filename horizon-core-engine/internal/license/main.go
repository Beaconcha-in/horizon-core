package main

import (
    "log"
    "horizon-core-engine/internal/license"
)

func main() {
    // ============================================================
    // 1. Initialize License Guard (executed first)
    // ============================================================
    guard := license.NewLicenseGuard()
    if err := guard.Start(); err != nil {
        log.Fatalf("❌ Startup halted: %v", err)
    }
    defer guard.Stop()

    // ============================================================
    // 2. Core Engine startup logic (only if license is valid)
    // ============================================================
    // Example: start HTTP server
    // if !guard.IsValid() {
    //     log.Fatal("❌ Invalid license. System halted.")
    // }

    // ... rest of Core Engine code ...
}