#!/bin/bash
# ============================================================
# 🚀 Horizon Core Engine – نصب و راه‌اندازی یکپارچه
# نسخه: 2.0.0
# توضیح: نصب Core Engine + راه‌اندازی سرور + دموی بانک مرکزی
# ============================================================

set -e  # توقف در صورت بروز خطا

# رنگ‌ها برای خروجی زیبا
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}🚀 Horizon Core Engine – نصب یکپارچه${NC}"
echo -e "${BLUE}============================================${NC}"

# ============================================================
# مرحله ۱: بررسی پیش‌نیازها
# ============================================================
echo -e "${YELLOW}📋 مرحله ۱: بررسی پیش‌نیازها...${NC}"

# بررسی وجود Git
if ! command -v git &> /dev/null; then
    echo -e "${RED}❌ Git نصب نیست. در حال نصب...${NC}"
    sudo apt-get update && sudo apt-get install -y git
else
    echo -e "${GREEN}✅ Git نصب است.${NC}"
fi

# بررسی وجود Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go نصب نیست. در حال نصب...${NC}"
    wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    rm go1.22.0.linux-amd64.tar.gz
    echo -e "${GREEN}✅ Go نصب شد.${NC}"
else
    echo -e "${GREEN}✅ Go نصب است.$(go version)${NC}"
fi

# بررسی وجود Node.js (برای داشبورد)
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js نصب نیست. در حال نصب...${NC}"
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt-get install -y nodejs
    echo -e "${GREEN}✅ Node.js نصب شد.${NC}"
else
    echo -e "${GREEN}✅ Node.js نصب است.$(node -v)${NC}"
fi

# ============================================================
# مرحله ۲: دریافت مخزن Horizon Core Engine
# ============================================================
echo -e "${YELLOW}📥 مرحله ۲: دریافت مخزن Horizon Core Engine...${NC}"

cd ~
if [ -d "horizon-core-engine" ]; then
    echo -e "${YELLOW}⚠️ پوشه horizon-core-engine وجود دارد. در حال به‌روزرسانی...${NC}"
    cd horizon-core-engine
    git pull origin main
else
    git clone https://github.com/beaconchain-horizon/horizon-core-engine.git
    cd horizon-core-engine
fi

echo -e "${GREEN}✅ مخزن دریافت شد.${NC}"

# ============================================================
# مرحله ۳: نصب وابستگی‌های Go
# ============================================================
echo -e "${YELLOW}📦 مرحله ۳: نصب وابستگی‌های Go...${NC}"
go mod tidy
echo -e "${GREEN}✅ وابستگی‌های Go نصب شد.${NC}"

# ============================================================
# مرحله ۴: تنظیم فایل محیطی (.env)
# ============================================================
echo -e "${YELLOW}⚙️ مرحله ۴: تنظیم فایل محیطی...${NC}"

if [ ! -f ".env" ]; then
    cp .env.example .env 2>/dev/null || echo "# Horizon Core Engine" > .env
    echo "PORT=8080" >> .env
    echo "DATABASE_URL=sqlite:./horizon.db" >> .env
    echo "CORE_API_URL=http://localhost:8080" >> .env
    echo -e "${GREEN}✅ فایل .env ایجاد شد.${NC}"
else
    echo -e "${GREEN}✅ فایل .env موجود است.${NC}"
fi

# ============================================================
# مرحله ۵: ساخت و اجرای Core Engine
# ============================================================
echo -e "${YELLOW}🔧 مرحله ۵: ساخت و اجرای Core Engine...${NC}"

# بررسی اینکه آیا پورت ۸۰۸۰ اشغال شده است
if command -v lsof &> /dev/null; then
    if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
        echo -e "${YELLOW}⚠️ پورت ۸۰۸۰ اشغال است. در حال توقف فرایند قبلی...${NC}"
        pkill -f "go run main.go" || true
        pkill -f "horizon-core" || true
        sleep 2
    fi
fi

# ساخت پروژه
go build -o horizon-core ./...
echo -e "${GREEN}✅ پروژه ساخته شد.${NC}"

# اجرا در پس‌زمینه
nohup ./horizon-core > logs.txt 2>&1 &
CORE_PID=$!
echo -e "${GREEN}✅ Core Engine با PID $CORE_PID اجرا شد.${NC}"
sleep 3

# بررسی اتصال
if curl -s http://localhost:8080/api/v1/status > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Core Engine پاسخ می‌دهد.${NC}"
else
    echo -e "${RED}❌ Core Engine پاسخ نمی‌دهد. لطفاً logs.txt را بررسی کنید.${NC}"
    exit 1
fi

# ============================================================
# مرحله ۶: استقرار داشبورد بانک مرکزی
# ============================================================
echo -e "${YELLOW}🌐 مرحله ۶: استقرار داشبورد بانک مرکزی...${NC}"

# ایجاد پوشه assets/js اگر وجود ندارد
mkdir -p assets/js

# ایجاد فایل dashboard.js
cat > assets/js/dashboard.js << 'EOF'
const CORE_API_URL = 'http://localhost:8080';

async function fetchValidatorsFromCore() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/validators`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        document.getElementById('statValidators').textContent = data.total?.toLocaleString() || '0';
        document.getElementById('statOnline').textContent = data.online?.toLocaleString() || '0';
        document.getElementById('statOffline').textContent = data.offline?.toLocaleString() || '0';
        document.getElementById('onlineCount').textContent = data.online?.toLocaleString() || '0';
        return data;
    } catch (error) {
        console.error('Error fetching validators:', error);
        return null;
    }
}

async function fetchTransactionsFromCore() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/transactions?limit=12`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        document.getElementById('statTx').textContent = data.total?.toLocaleString() || '0';
        const txContainer = document.getElementById('txList');
        if (txContainer && data.transactions && data.transactions.length > 0) {
            txContainer.innerHTML = data.transactions.map(tx => `
                <div class="tx-item">
                    <div><span class="tx-hash">${tx.hash?.slice(0, 12) || '...'}</span> 
                    <span class="tx-status">${tx.verified ? '✅' : '⏳'}</span></div>
                    <div style="font-size:0.5rem;color:#8fa2dc;">${tx.amount || '0'} ETH · ${new Date(tx.timestamp).toLocaleTimeString('en-US')}</div>
                </div>
            `).join('');
        }
        return data;
    } catch (error) {
        console.error('Error fetching transactions:', error);
        return null;
    }
}

async function fetchLicensesFromCore() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/licenses`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        document.getElementById('statLicenses').textContent = data.length?.toLocaleString() || '0';
        return data;
    } catch (error) {
        console.error('Error fetching licenses:', error);
        return null;
    }
}

async function fetchSystemStatus() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/status`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        document.getElementById('statApiCredit').textContent = data.api_credit || '0%';
        return data;
    } catch (error) {
        console.error('Error fetching system status:', error);
        return null;
    }
}

async function refreshAllData() {
    console.log('Fetching data from Core Engine...');
    await Promise.all([
        fetchValidatorsFromCore(),
        fetchTransactionsFromCore(),
        fetchLicensesFromCore(),
        fetchSystemStatus()
    ]);
    console.log('Data updated.');
}

let refreshInterval;

function startAutoRefresh() {
    refreshAllData();
    refreshInterval = setInterval(refreshAllData, 30000);
}

function stopAutoRefresh() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    startAutoRefresh();
});
EOF

echo -e "${GREEN}✅ dashboard.js ایجاد شد.${NC}"

# ============================================================
# مرحله ۷: دریافت فایل central-bank-demo.html
# ============================================================
echo -e "${YELLOW}📄 مرحله ۷: دریافت فایل داشبورد بانک مرکزی...${NC}"

curl -s -o central-bank-demo.html https://raw.githubusercontent.com/beaconchain-horizon/horizon-core-engine/main/central-bank-demo.html || echo -e "${YELLOW}⚠️ فایل central-bank-demo.html پیدا نشد. یک نسخه نمونه ایجاد می‌شود...${NC}"

# اگر فایل وجود نداشت، یک نسخه ساده ایجاد کن
if [ ! -f "central-bank-demo.html" ]; then
    cat > central-bank-demo.html << 'EOF'
<!DOCTYPE html>
<html lang="fa">
<head>
    <meta charset="UTF-8">
    <title>Horizon Core · بانک مرکزی</title>
</head>
<body>
    <h1>🏛️ Horizon Core Engine</h1>
    <p>داشبورد بانک مرکزی در حال بارگذاری...</p>
    <p>لطفاً فایل central-bank-demo.html را از مخزن دریافت کنید.</p>
</body>
</html>
EOF
fi

echo -e "${GREEN}✅ داشبورد بانک مرکزی آماده شد.${NC}"

# ============================================================
# مرحله ۸: نمایش اطلاعات نهایی
# ============================================================
echo -e "${BLUE}============================================${NC}"
echo -e "${GREEN}✅ نصب با موفقیت کامل شد!${NC}"
echo -e "${BLUE}============================================${NC}"
echo -e "${GREEN}🌐 Core Engine در آدرس زیر در حال اجراست:${NC}"
echo -e "${BLUE}   http://localhost:8080${NC}"
echo -e ""
echo -e "${GREEN}📊 داشبورد بانک مرکزی:${NC}"
echo -e "${BLUE}   http://localhost:8080/central-bank-demo.html${NC}"
echo -e ""
echo -e "${YELLOW}📝 لاگ‌ها:${NC}"
echo -e "   cat ~/horizon-core-engine/logs.txt"
echo -e ""
echo -e "${YELLOW}🛑 توقف Core Engine:${NC}"
echo -e "   pkill -f horizon-core"
echo -e ""
echo -e "${BLUE}============================================${NC}"
