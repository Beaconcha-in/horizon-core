#!/bin/bash
# Horizon Orchestrator - Fix nested structure and organize files

set -e

echo "========================================"
echo "🚀 Horizon Orchestrator - Fixing Structure"
echo "========================================"

# مرحله ۱: حذف پوشه‌ی تودرتو
if [ -d "horizon-core-engine" ] && [ -f "horizon-core-engine/go.mod" ]; then
    echo "⚠️ Nested 'horizon-core-engine' folder detected! Flattening..."
    cp -rf horizon-core-engine/* ./ 2>/dev/null || true
    cp -rf horizon-core-engine/.env ./ 2>/dev/null || true
    cp -rf horizon-core-engine/.gitignore ./ 2>/dev/null || true
    rm -rf horizon-core-engine
    echo "✅ Nested folder flattened."
else
    echo "✅ No nested structure detected."
fi

# مرحله ۲: ایجاد پوشه‌های استاندارد
mkdir -p cmd internal web scripts

# مرحله ۳: انتقال فایل‌ها به پوشه‌های مناسب
find . -maxdepth 1 -name "*.go" ! -name "*_test.go" | while read -r f; do
    if grep -q "package main" "$f" 2>/dev/null; then
        mv "$f" cmd/ 2>/dev/null || true
    else
        mv "$f" internal/ 2>/dev/null || true
    fi
done

mv *.html web/ 2>/dev/null || true
mv *.css web/ 2>/dev/null || true
mv *.js web/ 2>/dev/null || true
mv *.sh scripts/ 2>/dev/null || true

# مرحله ۴: تولید فایل‌های گمشده
if [ ! -f "go.mod" ]; then
    echo "⚠️ go.mod not found. Generating..."
    cat > go.mod <<'EOF'
module horizon

go 1.22

require (
    github.com/gin-contrib/cors v1.7.3
    github.com/gin-gonic/gin v1.10.0
    github.com/joho/godotenv v1.5.1
    gorm.io/driver/postgres v1.5.11
    gorm.io/gorm v1.25.12
)
EOF
fi

if [ ! -f "cmd/main.go" ]; then
    mkdir -p cmd
    echo "⚠️ main.go not found. Generating..."
    cat > cmd/main.go <<'EOF'
package main

import (
    "log"
    "os"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok", "message": "Horizon is alive!"})
    })
    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }
    log.Printf("Server running on port %s", port)
    r.Run(":" + port)
}
EOF
fi

echo "✅ Structure fixed successfully!"
echo "========================================"
echo "🎉 Operation completed successfully!"
