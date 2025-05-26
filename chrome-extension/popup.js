// Get elements
const statusIndicator = document.getElementById('statusIndicator');
const statusText = document.getElementById('statusText');
const autoRefreshToggle = document.getElementById('autoRefreshToggle');
const domainsInput = document.getElementById('domainsInput');
const saveDomainsBtn = document.getElementById('saveDomainsBtn');

// Load current status
chrome.runtime.sendMessage({ type: 'getStatus' }, (response) => {
  updateStatus(response.isConnected);
  autoRefreshToggle.checked = response.autoRefreshEnabled;
  domainsInput.value = response.enabledDomains.join('\n');
});

// Update UI status
function updateStatus(isConnected) {
  if (isConnected) {
    statusIndicator.classList.add('connected');
    statusIndicator.classList.remove('disconnected');
    statusText.textContent = 'Connected';
  } else {
    statusIndicator.classList.remove('connected');
    statusIndicator.classList.add('disconnected');
    statusText.textContent = 'Disconnected';
  }
}

// Handle auto-refresh toggle
autoRefreshToggle.addEventListener('change', (e) => {
  chrome.runtime.sendMessage({
    type: 'toggleAutoRefresh',
    enabled: e.target.checked
  });
});

// Handle save domains
saveDomainsBtn.addEventListener('click', () => {
  const domains = domainsInput.value
    .split('\n')
    .map(d => d.trim())
    .filter(d => d.length > 0);
    
  chrome.runtime.sendMessage({
    type: 'updateDomains',
    domains: domains
  }, () => {
    // Visual feedback
    saveDomainsBtn.textContent = 'Saved!';
    setTimeout(() => {
      saveDomainsBtn.textContent = 'Save Domains';
    }, 2000);
  });
});

