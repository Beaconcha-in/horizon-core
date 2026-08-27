#!/bin/bash

# Horizon Orchestrator - Auto Fix, Organize & Deploy
# This script will scan, fix, and deploy the entire Horizon ecosystem

set -e

REPO_ROOT=$(pwd)
LOG_FILE="$REPO_ROOT/orchestrator.log"

# ---------- رنگ‌بندی ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "$1" | tee -a "$LOG_FILE"; }

# ---------- مرحله ۱: اسکن ----------
log "${BLUE}🔍 Scanning project structure...${NC}"

# پیدا کردن همه فایل‌های Go
GO_FILES=$(find . -name "*.go" -type f)
HTML_FILES=$(find . -name "*.html" -type f)
CSS_FILES=$(find . -name "*.css" -type f)
JS_FILES=$(find . -name "*.js" -type f)
SH_FILES=$(find . -name "*.sh" -type f)

log "${GREEN}✅ Found ${#GO_FILES} Go files, ${#HTML_FILES} HTML files, ${#CSS_FILES} CSS files, ${#JS_FILES} JS files, ${#SH_FILES} Shell scripts${NC}"

# ---------- مرحله ۲: سازماندهی ----------
log "${BLUE}📁 Organizing files into standard structure...${NC}"

# ایجاد پوشه‌های اصلی
mkdir -p cmd internal pkg web scripts api

# انتقال فایل‌های Go به پوشه‌های مناسب
for f in $GO_FILES; do
    if [[ $f == *"main.go"* ]]; then
        mv "$f" cmd/ 2>/dev/null || true
    elif [[ $f == *"internal"* ]]; then
        mv "$f" internal/ 2>/dev/null || true
    else
        mv "$f" pkg/ 2>/dev/null || true
    fi
done

# انتقال فایل‌های وب
mv *.html web/ 2>/dev/null || true
mv *.css web/ 2>/dev/null || true
mv *.js web/ 2>/dev/null || true

# انتقال اسکریپت‌ها
mv *.sh scripts/ 2>/dev/null || true

log "${GREEN}✅ Files organized.${NC}"

# ---------- مرحله ۳: تولید فایل‌های گمشده ----------
log "${BLUE}⚙️ Checking for missing essential files...${NC}"

# تولید go.mod اگر وجود نداشت
if [ ! -f "go.mod" ]; then
    log "${YELLOW}⚠️ go.mod not found. Generating...${NC}"
    cat > go.mod <<EOF
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

# تولید liara.json اگر وجود نداشت
if [ ! -f "liara.json" ]; then
    log "${YELLOW}⚠️ liara.json not found. Generating...${NC}"
    cat > liara.json <<EOF
{
  "app": "horizon-core",
  "port": 3000,
  "build": {
    "location": "iran",
    "builder": "go",
    "command": "go build -o horizon ./cmd && ./horizon"
  },
  "disks": [],
  "env": {
    "GO_VERSION": "1.22"
  }
}
EOF
fi

# تولید main.go اگر وجود نداشت
if [ ! -f "cmd/main.go" ]; then
    log "${YELLOW}⚠️ main.go not found in cmd/. Generating...${NC}"
    mkdir -p cmd
    cat > cmd/main.go <<EOF
package main

import (
    "log"
    "os"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
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

log "${GREEN}✅ All essential files are in place.${NC}"

# ---------- مرحله ۴: اتصال به GitHub API و هماهنگی مخازن ----------
log "${BLUE}🌐 Syncing with GitHub repositories...${NC}"

GITHUB_TOKEN=${GITHUB_TOKEN:-""}
if [ -n "$GITHUB_TOKEN" ]; then
    # دریافت لیست مخازن موجود
    REPOS=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
        "https://api.github.com/user/repos?per_page=100" | jq -r '.[].name')

    # لیست مخازن مورد نیاز
    REQUIRED_REPOS=("horizon-core-engine" "horizon-license-server" "horizon-switch" "horizon-dashboard" "horizon-dex")

    for repo in "${REQUIRED_REPOS[@]}"; do
        if echo "$REPOS" | grep -q "^$repo$"; then
            log "${GREEN}✅ Repository $repo already exists.${NC}"
        else
            log "${YELLOW}⚠️ Repository $repo not found. Creating...${NC}"
            curl -X POST -H "Authorization: token $GITHUB_TOKEN" \
                -d "{\"name\":\"$repo\", \"private\":false, \"auto_init\":true}" \
                "https://api.github.com/user/repos"
            log "${GREEN}✅ Repository $repo created.${NC}"
        fi
    done
else
    log "${YELLOW}⚠️ GITHUB_TOKEN not set. Skipping GitHub repo sync.${NC}"
fi

# ---------- مرحله ۵: نصب وابستگی‌ها و استقرار ----------
log "${BLUE}📦 Installing dependencies and deploying...${NC}"

go mod tidy
npm install -g @liara/cli

if [ -n "$LIARA_TOKEN" ]; then
    liara deploy --api-token="$LIARA_TOKEN"
    log "${GREEN}✅ Deployment to Liara completed.${NC}"
else
    log "${YELLOW}⚠️ LIARA_TOKEN not set. Skipping deployment.${NC}"
fi

log "${BLUE}🎯 Orchestration completed successfully!${NC}"
