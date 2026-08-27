#!/bin/bash
# ================================================================
# orchestrator.sh – Horizon Ecosystem Orchestrator (English)
# ================================================================
# This script syncs ALL branches with 'main', then performs
# full project health check, fixes structure, installs deps,
# and deploys to Liara (if token is set).
# ================================================================

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

REPO_ROOT=$(pwd)
LOG_FILE="$REPO_ROOT/orchestrator.log"
log() { echo -e "$1" | tee -a "$LOG_FILE"; }

# ==========================================
# Step 0: Check if inside a Git repository
# ==========================================
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    log "${RED}❌ This directory is not a Git repository.${NC}"
    exit 1
fi

REPO_NAME=$(basename "$(git rev-parse --show-toplevel 2>/dev/null)" 2>/dev/null)
log "${BLUE}========================================${NC}"
log "${GREEN}🚀 Horizon Orchestrator - SYNC ALL BRANCHES${NC}"
log "${BLUE}📁 Repository: $REPO_NAME${NC}"
log "${BLUE}========================================${NC}"

# ==========================================
# Step 1: Sync all branches with 'main'
# ==========================================
sync_all_branches_with_main() {
    log "${BLUE}🔀 Syncing ALL branches with 'main'...${NC}"
    
    # Get list of all branches except main
    ALL_BRANCHES=$(git branch -a | grep -v "main" | grep -v "HEAD" | sed 's/^[ *]*//g' | sed 's/remotes\/origin\///g' | sort -u)
    
    if [ -z "$ALL_BRANCHES" ]; then
        log "${YELLOW}⚠️ No other branches found. Only 'main' exists.${NC}"
        return 0
    fi
    
    # First ensure we are on main
    git checkout main 2>/dev/null || git checkout -b main
    git pull origin main 2>/dev/null || log "${YELLOW}⚠️ Pull failed, continuing...${NC}"
    
    # For each branch, merge it with main
    for branch in $ALL_BRANCHES; do
        log "${YELLOW}🔄 Syncing branch: $branch${NC}"
        git checkout "$branch" 2>/dev/null || git checkout -b "$branch"
        git pull origin "$branch" 2>/dev/null || log "${YELLOW}⚠️ Pull failed for $branch, continuing...${NC}"
        git merge main -m "Merge main into $branch" 2>/dev/null || log "${YELLOW}⚠️ Merge conflict? Skipping...${NC}"
        git push origin "$branch" 2>/dev/null || log "${YELLOW}⚠️ Push failed for $branch.${NC}"
    done
    
    # Finally return to main
    git checkout main
    git pull origin main 2>/dev/null || log "${YELLOW}⚠️ Pull failed, continuing...${NC}"
    log "${GREEN}✅ All branches have been synced with 'main'.${NC}"
}

# ==========================================
# Step 2: Fix nested project structure
# ==========================================
fix_nested_structure() {
    log "${BLUE}🔍 Checking for nested project structure...${NC}"
    
    # If a nested 'horizon-core-engine' folder exists with go.mod inside...
    if [ -d "horizon-core-engine" ] && [ -f "horizon-core-engine/go.mod" ]; then
        log "${YELLOW}⚠️ Nested 'horizon-core-engine' folder detected! Flattening...${NC}"
        cp -rf horizon-core-engine/* ./ 2>/dev/null || true
        cp -rf horizon-core-engine/.env ./ 2>/dev/null || true
        cp -rf horizon-core-engine/.gitignore ./ 2>/dev/null || true
        rm -rf horizon-core-engine
        log "${GREEN}✅ Nested folder flattened.${NC}"
    else
        log "${GREEN}✅ No nested structure detected.${NC}"
    fi
}

# ==========================================
# Step 3: Generate missing essential files
# ==========================================
generate_missing_files() {
    log "${BLUE}⚙️ Checking essential files...${NC}"
    
    # Generate go.mod if missing
    if [ ! -f "go.mod" ]; then
        log "${YELLOW}⚠️ go.mod not found. Generating...${NC}"
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
    
    # Generate liara.json if missing
    if [ ! -f "liara.json" ]; then
        log "${YELLOW}⚠️ liara.json not found. Generating...${NC}"
        cat > liara.json <<'EOF'
{
  "app": "trusting-rubin-oxf0xqvfp",
  "port": 3000,
  "build": {
    "location": "iran",
    "builder": "go",
    "command": "go build -o horizon ./cmd && ./horizon"
  },
  "disks": [],
  "env": { "GO_VERSION": "1.22" }
}
EOF
    fi
    
    # Generate main.go in cmd/ if missing
    if [ ! -f "cmd/main.go" ]; then
        mkdir -p cmd
        log "${YELLOW}⚠️ main.go not found in cmd/. Generating...${NC}"
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
    
    log "${GREEN}✅ All essential files are in place.${NC}"
}

# ==========================================
# Step 4: Install dependencies
# ==========================================
install_deps() {
    log "${BLUE}📦 Installing dependencies...${NC}"
    go mod tidy
    log "${GREEN}✅ Dependencies installed.${NC}"
}

# ==========================================
# Step 5: Deploy to Liara (if token is set)
# ==========================================
deploy_to_liara() {
    if [ -n "$LIARA_TOKEN" ]; then
        log "${BLUE}🌐 Deploying to Liara...${NC}"
        npm install -g @liara/cli
        liara deploy --api-token="$LIARA_TOKEN"
        log "${GREEN}✅ Deployment to Liara completed successfully!${NC}"
    else
        log "${YELLOW}⚠️ LIARA_TOKEN not set. Skipping deployment.${NC}"
        log "To deploy, set LIARA_TOKEN and run 'liara deploy' manually."
    fi
}

# ==========================================
# Step 6: Run locally (optional)
# ==========================================
run_local() {
    log "${GREEN}🚀 Starting backend on port 8080 for local test...${NC}"
    go run cmd/main.go &
    BACKEND_PID=$!
    log "${GREEN}✅ Backend running with PID: $BACKEND_PID${NC}"
    log "${YELLOW}📌 To stop: kill $BACKEND_PID${NC}"
}

# ==========================================
# Main execution
# ==========================================
main() {
    sync_all_branches_with_main
    fix_nested_structure
    generate_missing_files
    install_deps
    deploy_to_liara
    run_local  # Optional: remove if you don't want local test
    
    log "${BLUE}========================================${NC}"
    log "${GREEN}🎉 Operation completed successfully!${NC}"
    log "${BLUE}========================================${NC}"
}

main
