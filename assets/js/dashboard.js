const CORE_API_URL = 'https://horizon-core-engine.onrender.com';

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
