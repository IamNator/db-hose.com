const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080';

// Helper function to handle HTTP requests
async function makeRequest(endpoint, method, data = null, token = null) {
    const headers = {
        'Content-Type': 'application/json',
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        method: method,
        headers: headers,
        body: data ? JSON.stringify(data) : null,
    });

    if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Request failed');
    }

    return response.json();
}

// API Functions
async function signup(email, password) {
    return makeRequest('/signup', 'POST', { email, password });
}

async function login(email, password) {
    const response = await makeRequest('/login', 'POST', { email, password });
    store.set('jwtToken', response.token); // Store the JWT token using Store.js
    return response;
}

async function storeCreds(secret, creds) {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest(`/credentials/store?secret=${encodeURIComponent(secret)}`, 'POST', creds, token);
}

async function editCreds(secret, creds) {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest(`/credentials/edit?secret=${encodeURIComponent(secret)}`, 'PUT', creds, token);
}

async function deleteCreds(key) {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest(`/credentials/delete/${key}`, 'DELETE', null, token);
}

async function viewCreds(key, secret) {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest(`/credentials/view/${key}?secret=${encodeURIComponent(secret)}`, 'GET', null, token);
}

async function listCreds() {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest('/credentials/list', 'GET', null, token);
}

async function backup(key, secret) {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest(`/backup/${key}?secret=${encodeURIComponent(secret)}`, 'POST', null, token);
}

async function restore(key, secret, fileName) {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest(`/restore/${key}?secret=${encodeURIComponent(secret)}&file=${encodeURIComponent(fileName)}`, 'POST', null, token);
}

async function fetchLogs() {
    const token = store.get('jwtToken'); // Retrieve the JWT token from Store.js
    return makeRequest('/logs', 'GET', null, token);
}
