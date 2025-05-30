let websocket = null;
let isConnected = false;
let reconnectInterval = null;
let keepAliveInterval = null;
let enabledDomains = ['localhost', '127.0.0.1', 'LayeredApps/*'];
let autoRefreshEnabled = true;

// Load saved settings on startup
chrome.storage.local.get(['enabledDomains', 'autoRefreshEnabled'], (result) => {
  if (result.enabledDomains) {
    enabledDomains = result.enabledDomains;
  }
  if (result.autoRefreshEnabled !== undefined) {
    autoRefreshEnabled = result.autoRefreshEnabled;
  }
});

// Sanitize Windows paths in JSON data
function sanitizeWindowsPaths(jsonData) {
  // Handle unescaped backslashes in file paths
  if (jsonData.includes('"filename":"') && jsonData.includes('\\')) {
    jsonData = jsonData.replace(/"filename":"([^"]*?)"/g, (match, filename) => {
      const escapedFilename = filename.replace(/\\/g, '\\\\');
      return `"filename":"${escapedFilename}"`;
    });
  }

  return jsonData;
}

// Connect to WebSocket server
function connectWebSocket() {
  if (websocket && websocket.readyState === WebSocket.OPEN) {
    return;
  }

  websocket = new WebSocket('ws://localhost:8080/ws');

  websocket.onopen = () => {
    console.log('Connected to Layered Code WebSocket server');
    isConnected = true;
    chrome.action.setBadgeText({ text: ' ' });
    chrome.action.setBadgeBackgroundColor({ color: '#4CAF50' });

    // Clear reconnect interval if exists
    if (reconnectInterval) {
      clearInterval(reconnectInterval);
      reconnectInterval = null;
    }

    // Set up keep-alive ping every 30 seconds
    if (keepAliveInterval) {
      clearInterval(keepAliveInterval);
    }
    keepAliveInterval = setInterval(() => {
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        websocket.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);
  };

  websocket.onmessage = (event) => {
    try {
      const jsonData = sanitizeWindowsPaths(event.data);
      const data = JSON.parse(jsonData);

      if (data.type === 'file-changed' && autoRefreshEnabled) {
        console.log('File changed:', data.filename, 'Action:', data.action);
        refreshMatchingTabs(data.filename);
      }
    } catch (err) {
      console.error('Error parsing WebSocket message:', err);
    }
  };

  websocket.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  websocket.onclose = (event) => {
    console.log('Disconnected from Layered Code WebSocket server', event.code, event.reason);
    isConnected = false;
    chrome.action.setBadgeText({ text: ' ' });
    chrome.action.setBadgeBackgroundColor({ color: '#F44336' });

    // Clean up keep-alive interval
    if (keepAliveInterval) {
      clearInterval(keepAliveInterval);
      keepAliveInterval = null;
    }

    // Clean up the closed websocket
    websocket = null;

    // Attempt to reconnect every 2 seconds (faster reconnection)
    if (!reconnectInterval) {
      reconnectInterval = setInterval(() => {
        console.log('Attempting to reconnect...');
        connectWebSocket();
      }, 2000);
    }
  };
}

// Track recently refreshed tabs to prevent loops
let recentlyRefreshedTabs = new Set();
let refreshDebounceTimer = null;

// Clear recently refreshed tabs after a delay
function clearRecentlyRefreshed(tabId) {
  setTimeout(() => {
    recentlyRefreshedTabs.delete(tabId);
  }, 2000); // Clear after 2 seconds
}

// Check if a URL matches any enabled domain
function shouldRefreshUrl(url) {
  const urlLower = url.toLowerCase();

  return enabledDomains.some(domain => {
    const domainLower = domain.toLowerCase();

    // Handle wildcard patterns
    if (domain.includes('*')) {
      const pattern = domainLower.replace(/\*/g, '.*');
      const regex = new RegExp(pattern, 'i');
      return regex.test(urlLower);
    }

    // Simple string matching for all other cases
    return urlLower.includes(domainLower);
  });
}

// Refresh tabs that match the enabled domains
async function refreshMatchingTabs(filename) {
  try {
    // Clear any existing debounce timer
    if (refreshDebounceTimer) {
      clearTimeout(refreshDebounceTimer);
    }

    // Debounce rapid file changes
    refreshDebounceTimer = setTimeout(async () => {
      const tabs = await chrome.tabs.query({});
      console.log(`Checking ${tabs.length} tabs for refresh...`);

      for (const tab of tabs) {
        if (!tab.url) continue;

        // Skip if tab was recently refreshed
        if (recentlyRefreshedTabs.has(tab.id)) {
          console.log(`Skipping recently refreshed tab: ${tab.id}`);
          continue;
        }

        if (shouldRefreshUrl(tab.url)) {
          console.log(`Refreshing tab: ${tab.title} (${tab.url})`);
          recentlyRefreshedTabs.add(tab.id);
          clearRecentlyRefreshed(tab.id);
          await chrome.tabs.reload(tab.id);
        }
      }
    }, 100); // 100ms debounce
  } catch (err) {
    console.error('Error refreshing tabs:', err);
  }
}

// Listen for messages from popup
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'getStatus') {
    sendResponse({
      isConnected: isConnected,
      autoRefreshEnabled: autoRefreshEnabled,
      enabledDomains: enabledDomains
    });
  } else if (request.type === 'toggleAutoRefresh') {
    autoRefreshEnabled = request.enabled;
    chrome.storage.local.set({ autoRefreshEnabled: autoRefreshEnabled });
    sendResponse({ success: true });
  } else if (request.type === 'updateDomains') {
    enabledDomains = request.domains;
    chrome.storage.local.set({ enabledDomains: enabledDomains });
    sendResponse({ success: true });
  }
  return true;
});

// Initialize connection
connectWebSocket();

// Ensure connection on extension startup/install
chrome.runtime.onStartup.addListener(() => {
  console.log('Extension started, connecting to WebSocket...');
  connectWebSocket();
});

chrome.runtime.onInstalled.addListener(() => {
  console.log('Extension installed/updated, connecting to WebSocket...');
  connectWebSocket();
});

// Listen for tab updates to handle navigation
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete' && tab.url) {
    // Clear the recently refreshed status for this tab after navigation completes
    if (recentlyRefreshedTabs.has(tabId)) {
      setTimeout(() => {
        recentlyRefreshedTabs.delete(tabId);
      }, 1000);
    }
  }
});