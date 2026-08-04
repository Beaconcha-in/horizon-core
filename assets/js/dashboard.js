// =============================================================
// 🔌 CONNECTION TO CORE ENGINE
// =============================================================

const CORE_API_URL = 'https://horizon-core-engine.onrender.com';

// =============================================================
// 📡 FETCH REAL DATA FROM CORE ENGINE
// =============================================================

/**
 * Fetch validators list from Core Engine
 */
async function fetchValidatorsFromCore() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/validators`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        
        // Update DOM elements (with existence check)
        const elements = {
            statValidators: document.getElementById('statValidators'),
            statOnline: document.getElementById('statOnline'),
            statOffline: document.getElementById('statOffline'),
            onlineCount: document.getElementById('onlineCount')
        };
        
        if (elements.statValidators) elements.statValidators.textContent = data.total?.toLocaleString() || '0';
        if (elements.statOnline) elements.statOnline.textContent = data.online?.toLocaleString() || '0';
        if (elements.statOffline) elements.statOffline.textContent = data.offline?.toLocaleString() || '0';
        if (elements.onlineCount) elements.onlineCount.textContent = data.online?.toLocaleString() || '0';
        
        return data;
    } catch (error) {
        console.error('❌ Error fetching validators:', error);
        showConnectionError('Validators');
        return null;
    }
}

/**
 * Fetch recent transactions from Core Engine
 */
async function fetchTransactionsFromCore() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/transactions?limit=12`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        
        const statTx = document.getElementById('statTx');
        if (statTx) statTx.textContent = data.total?.toLocaleString() || '0';
        
        const txContainer = document.getElementById('txList');
        if (txContainer) {
            const transactions = data.transactions || [];
            if (transactions.length > 0) {
                txContainer.innerHTML = transactions.map(tx => `
                    <div class="tx-item">
                        <div>
                            <span class="tx-hash">${tx.hash?.slice(0, 12) || '...'}</span> 
                            <span class="tx-status">${tx.verified ? '✅' : '⏳'}</span>
                        </div>
                        <div style="font-size:0.5rem;color:#8fa2dc;">
                            ${tx.amount || '0'} ETH · ${new Date(tx.timestamp).toLocaleTimeString('en-US')}
                        </div>
                    </div>
                `).join('');
            } else {
                txContainer.innerHTML = '<div class="tx-item">No transactions available</div>';
            }
        }
        return data;
    } catch (error) {
        console.error('❌ Error fetching transactions:', error);
        showConnectionError('Transactions');
        return null;
    }
}

/**
 * Fetch issued licenses from Core Engine
 */
async function fetchLicensesFromCore() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/licenses`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        
        const statLicenses = document.getElementById('statLicenses');
        if (statLicenses) statLicenses.textContent = data.length?.toLocaleString() || '0';
        
        return data;
    } catch (error) {
        console.error('❌ Error fetching licenses:', error);
        showConnectionError('Licenses');
        return null;
    }
}

/**
 * Fetch system status from Core Engine
 */
async function fetchSystemStatus() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/status`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        
        const statApiCredit = document.getElementById('statApiCredit');
        if (statApiCredit) statApiCredit.textContent = data.api_credit || '0%';
        
        return data;
    } catch (error) {
        console.error('❌ Error fetching system status:', error);
        showConnectionError('System Status');
        return null;
    }
}

/**
 * Show connection error in the panel
 */
function showConnectionError(service) {
    const statusEl = document.getElementById('statApiCredit');
    if (statusEl) {
        statusEl.textContent = '❌ Offline';
        statusEl.style.color = '#f87171';
    }
    console.warn(`⚠️ Connection to ${service} service failed.`);
}

// =============================================================
// 🔄 MAIN FUNCTION TO REFRESH ALL DATA
// =============================================================

async function refreshAllData() {
    console.log('🔄 Fetching data from Core Engine...');
    try {
        await Promise.all([
            fetchValidatorsFromCore(),
            fetchTransactionsFromCore(),
            fetchLicensesFromCore(),
            fetchSystemStatus()
        ]);
        console.log('✅ Data updated successfully.');
    } catch (error) {
        console.error('❌ Error during data refresh:', error);
    }
}

// =============================================================
// ⏱️ AUTO-REFRESH TIMER
// =============================================================

let refreshInterval;

function startAutoRefresh() {
    // Run immediately on first call
    refreshAllData();
    
    // Then every 30 seconds
    refreshInterval = setInterval(refreshAllData, 30000);
}

function stopAutoRefresh() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
        console.log('🛑 Auto-refresh stopped.');
    }
}

// =============================================================
// 🧪 CONNECTION TEST
// =============================================================

async function testConnection() {
    try {
        const response = await fetch(`${CORE_API_URL}/api/v1/status`, {
            method: 'HEAD',
            signal: AbortSignal.timeout(5000)
        });
        if (response.ok) {
            console.log('✅ Connection to Core Engine established.');
            return true;
        }
    } catch (error) {
        console.error('❌ Cannot connect to Core Engine:', error.message);
        return false;
    }
}

// =============================================================
// 🚀 AUTO-START ON PAGE LOAD
// =============================================================

document.addEventListener('DOMContentLoaded', async () => {
    console.log('🚀 Dashboard initializing...');
    
    // Test connection
    const connected = await testConnection();
    
    if (connected) {
        startAutoRefresh();
    } else {
        console.warn('⚠️ Data is not live. Please start Core Engine.');
        // Show offline message in the panel
        const statusEl = document.getElementById('statApiCredit');
        if (statusEl) {
            statusEl.textContent = '❌ Offline';
            statusEl.style.color = '#f87171';
        }
        // Retry after 10 seconds
        setTimeout(async () => {
            const reconnected = await testConnection();
            if (reconnected) {
                startAutoRefresh();
            }
        }, 10000);
    }
});

// =============================================================
// 🛑 EXPOSE STOP FUNCTION FOR EMERGENCY SHUTDOWN
// =============================================================
window.stopAutoRefresh = stopAutoRefresh;

console.log('📡 Dashboard.js loaded successfully.');
