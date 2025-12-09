const API_BASE = '/api';

// Session management
let refreshTokenTimer = null;

// Show different forms
function showLogin() {
    hideAllForms();
    document.getElementById('login-form').style.display = 'block';
}

function showRegister() {
    hideAllForms();
    document.getElementById('register-form').style.display = 'block';
}

function showResetRequest() {
    hideAllForms();
    document.getElementById('reset-request-form').style.display = 'block';
}

function showResetConfirm(token) {
    hideAllForms();
    document.getElementById('reset-confirm-form').style.display = 'block';
    document.getElementById('reset-token').value = token;
}

function hideAllForms() {
    document.querySelectorAll('.auth-form').forEach(form => {
        form.style.display = 'none';
    });
    hideMessage();
}

// Message display
function showMessage(message, type = 'success') {
    const messageEl = document.getElementById('auth-message');
    messageEl.textContent = message;
    messageEl.className = type;
    messageEl.style.display = 'block';
}

function hideMessage() {
    document.getElementById('auth-message').style.display = 'none';
}

// Handle registration
async function handleRegister(e) {
    e.preventDefault();

    const email = document.getElementById('register-email').value;
    const fullName = document.getElementById('register-name').value;
    const password = document.getElementById('register-password').value;
    const confirm = document.getElementById('register-confirm').value;

    if (password !== confirm) {
        showMessage('Las contraseñas no coinciden', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password, full_name: fullName })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Error al registrar');
        }

        // Store tokens
        saveSession(data);

        showMessage('¡Cuenta creada exitosamente! Redirigiendo...', 'success');
        setTimeout(() => {
            window.location.href = '/';
        }, 1500);
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// Handle login
async function handleLogin(e) {
    e.preventDefault();

    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    const rememberMe = document.getElementById('remember-me').checked;

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password, remember_me: rememberMe })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Error al iniciar sesión');
        }

        // Store tokens
        saveSession(data);

        showMessage('¡Inicio de sesión exitoso! Redirigiendo...', 'success');
        setTimeout(() => {
            window.location.href = '/';
        }, 1500);
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// Handle password reset request
async function handleResetRequest(e) {
    e.preventDefault();

    const email = document.getElementById('reset-email').value;

    try {
        const response = await fetch(`${API_BASE}/auth/request-reset`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Error al solicitar restablecimiento');
        }

        showMessage(data.message, 'success');
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// Handle password reset confirmation
async function handleResetConfirm(e) {
    e.preventDefault();

    const token = document.getElementById('reset-token').value;
    const newPassword = document.getElementById('new-password').value;
    const confirm = document.getElementById('new-password-confirm').value;

    if (newPassword !== confirm) {
        showMessage('Las contraseñas no coinciden', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/auth/reset-password`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ token, new_password: newPassword })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Error al restablecer contraseña');
        }

        showMessage('Contraseña restablecida exitosamente. Redirigiendo...', 'success');
        setTimeout(() => {
            showLogin();
        }, 1500);
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// Session management functions
function saveSession(data) {
    localStorage.setItem('access_token', data.token);
    localStorage.setItem('refresh_token', data.refresh_token);
    localStorage.setItem('user', JSON.stringify(data.user));
    localStorage.setItem('expires_in', data.expires_in);
    localStorage.setItem('login_time', Date.now());

    // Schedule token refresh
    scheduleTokenRefresh(data.expires_in);
}

function clearSession() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    localStorage.removeItem('expires_in');
    localStorage.removeItem('login_time');

    if (refreshTokenTimer) {
        clearTimeout(refreshTokenTimer);
        refreshTokenTimer = null;
    }
}

function getAccessToken() {
    return localStorage.getItem('access_token');
}

function getRefreshToken() {
    return localStorage.getItem('refresh_token');
}

function isLoggedIn() {
    return !!getAccessToken();
}

function getCurrentUser() {
    const userJson = localStorage.getItem('user');
    return userJson ? JSON.parse(userJson) : null;
}

// Token refresh logic
function scheduleTokenRefresh(expiresIn) {
    // Refresh token 5 minutes before expiration
    const refreshTime = (expiresIn - 300) * 1000;

    if (refreshTime > 0) {
        refreshTokenTimer = setTimeout(async () => {
            await refreshAccessToken();
        }, refreshTime);
    }
}

async function refreshAccessToken() {
    const refreshToken = getRefreshToken();
    if (!refreshToken) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/auth/refresh`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ refresh_token: refreshToken })
        });

        if (response.ok) {
            const data = await response.json();
            saveSession(data);
        } else {
            // Refresh failed, clear session
            clearSession();
            window.location.href = '/auth.html';
        }
    } catch (error) {
        console.error('Error refreshing token:', error);
        clearSession();
        window.location.href = '/auth.html';
    }
}

// OAuth2 functions
async function checkOAuth2Config() {
    try {
        const response = await fetch(`${API_BASE}/oauth2/config`);
        const data = await response.json();

        if (data.enabled) {
            document.getElementById('oauth2-login').style.display = 'block';
        }
    } catch (error) {
        console.error('Error checking OAuth2 config:', error);
    }
}

async function handleOAuth2Login() {
    try {
        const response = await fetch(`${API_BASE}/oauth2/login`);
        const data = await response.json();

        if (data.auth_url) {
            // Redirect to OAuth2 provider
            window.location.href = data.auth_url;
        } else {
            showMessage('Error al iniciar sesión con Microsoft', 'error');
        }
    } catch (error) {
        showMessage(error.message, 'error');
    }
}

// Check for OAuth2 callback
async function handleOAuth2Callback() {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state');
    const error = params.get('error');

    if (error) {
        showMessage('Error de autenticación: ' + (params.get('error_description') || error), 'error');
        showLogin();
        return;
    }

    if (code && state) {
        showMessage('Autenticando...', 'success');

        try {
            // The callback endpoint will handle the OAuth2 exchange
            const response = await fetch(`${API_BASE}/oauth2/callback?code=${code}&state=${state}`);
            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Error al autenticar');
            }

            // Store tokens
            saveSession(data);

            showMessage('¡Autenticación exitosa! Redirigiendo...', 'success');
            setTimeout(() => {
                window.location.href = '/';
            }, 1500);
        } catch (error) {
            showMessage(error.message, 'error');
            setTimeout(() => {
                window.location.href = '/auth.html';
            }, 2000);
        }
        return true;
    }

    return false;
}

// Check for reset token in URL or OAuth2 callback
window.addEventListener('DOMContentLoaded', async () => {
    // Check for OAuth2 callback first
    const isCallback = await handleOAuth2Callback();
    if (isCallback) return;

    const params = new URLSearchParams(window.location.search);
    const resetToken = params.get('token');

    if (resetToken) {
        showResetConfirm(resetToken);
    } else {
        showLogin();
        // Check if OAuth2 is available
        await checkOAuth2Config();
    }
});
