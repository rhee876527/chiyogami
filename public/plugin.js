function showToast(message, type = 'info') {
    const existingToast = document.querySelector('.toast');
    if (existingToast) {
        existingToast.remove();
    }
    const toast = document.createElement('div');
    toast.className = `z-50 toast fixed bottom-4 right-4 p-4 rounded-md text-white ${type === 'error' ? 'bg-red-500' : 'bg-green-500'}`;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 6000);
}

async function fetchWithAuth(url, options = {}) {
    try {
        const response = await fetch(url, options);

        if (response.status === 401) {
            updateAuthUI(null);
            return null; //silence unauthorized errors
        }

        return response;
    } catch (error) {
        console.error("Fetch error:", error);
        throw error;
    }
}

let deleteAccountButton;

async function checkAuthStatus() {
    if (!localStorage.getItem('username')) return updateAuthUI(null);

    const response = await fetchWithAuth('/user/pastes');
    if (!response?.ok) {
        updateAuthUI(null);
        return;
    }

    const data = await response.json().catch(() => {});
    if (data?.username) updateAuthUI(data.username);
}

function updateAuthUI(username) {
    const authButtons = document.getElementById('auth-buttons');
    const userInfo = document.getElementById('user-info');
    const usernameDisplay = document.getElementById('username-display');

    if (username) {
        authButtons.classList.add('hidden');
        userInfo.classList.remove('hidden');
        usernameDisplay.textContent = username;
        if (deleteAccountButton) deleteAccountButton.classList.remove('hidden');
    } else {
        authButtons.classList.remove('hidden');
        userInfo.classList.add('hidden');
        usernameDisplay.textContent = '';
        if (deleteAccountButton) deleteAccountButton.classList.add('hidden');
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const createButton = document.querySelector('.create-button');
    const cancelButton = document.querySelector('.cancel-button');
    const loginButton = document.getElementById('login-button');
    const registerButton = document.getElementById('register-button');
    const logoutButton = document.getElementById('logout-button');
    const myPastesButton = document.getElementById('my-pastes-button');

    const loginModal = document.getElementById('login-modal');
    const registerModal = document.getElementById('register-modal');
    const myPastesModal = document.getElementById('my-pastes-modal');

    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');

    const deleteAccountButton = document.getElementById('delete-account-button');
    const deleteAccountModal = document.getElementById('delete-account-modal');
    const confirmDeleteAccount = document.getElementById('confirm-delete-account');
    const cancelDeleteAccount = document.getElementById('cancel-delete-account');

    deleteAccountButton.addEventListener('click', () => showModal(deleteAccountModal));
    cancelDeleteAccount.addEventListener('click', () => hideModal(deleteAccountModal));

    confirmDeleteAccount.addEventListener('click', async () => {
        try {
            const response = await fetch('/delete-account', {
                method: 'DELETE',
                credentials: 'same-origin'
            });
            if (response.ok) {
                updateAuthUI(null);
                hideModal(deleteAccountModal);
                showToast('Account deleted successfully', 'info');
            } else {
                const errorText = await response.json();
                throw new Error(`Failed to delete account: ${errorText.message}`);
            }
        } catch (error) {
            console.error('Delete account error:', error);
            showToast('Unable to connect to server', 'error');
        }
    });

    function showModal(modal) {
        modal.classList.add('modal-open');
    }

    function hideModal(modal) {
        modal.classList.remove('modal-open');
    }

    loginButton.addEventListener('click', () => showModal(loginModal));
    registerButton.addEventListener('click', () => showModal(registerModal));

    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('login-username').value;
        const password = document.getElementById('login-password').value;

        try {
            const response = await fetch('/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
            });

            if (response.ok) {
                updateAuthUI(username);
                localStorage.setItem('username', username);
                hideModal(loginModal);
                showToast('Login successful', 'info');
            } else {
                const badLogin = await response.json();
                showToast(`Login failed: ${badLogin.message}`, 'error');
            }
        } catch (error) {
            console.error('Login Error:', error);
            showToast('Unable to connect to server', 'error');
        }
    });

    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('register-username').value;
        const password = document.getElementById('register-password').value;

        try {
            const response = await fetch('/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username, password }),
            });

            if (response.ok) {
                showToast('Registration successful. Please log in.', 'info');
                hideModal(registerModal);
            } else {
                const errorData = await response.json();
                showToast(errorData.message.includes('UNIQUE constraint failed') ? 'Username already taken' : `${errorData.message}`, 'error');
            }
        } catch (error) {
            console.error('Registration error:', error);
            showToast('Unable to connect to server', 'error');
        }
    });

    logoutButton.addEventListener('click', async () => {
        try {
            const response = await fetch('/logout', { method: 'POST' });
            if (response.ok) {
                localStorage.removeItem('username');
                updateAuthUI(null);
                showToast('Logout successful', 'info');
            }
        } catch (error) {
            console.error('Logout error:', error);
            showToast('Unable to connect to server', 'error');
        }
    });

    myPastesButton.addEventListener("click", async () => {
        try {
            const response = await fetchWithAuth('/user/pastes');

            if (!response.ok) {
                if (response.status === 401) {
                    showToast("Session expired. Please log in again.", "error");
                    return;
                }
                throw new Error("Failed to fetch pastes");
            }

            const pastes = await response.json();
            const pastesList = document.getElementById("my-pastes-list");
            pastesList.innerHTML = "";

            if (!pastes.length) {
                pastesList.innerHTML = "No pastes available.";
            } else {
                pastes.reverse().forEach((paste) => {
                    const li = document.createElement("li");
                    li.className = "flex justify-between items-center mb-1.5";
                    li.innerHTML = `
                        <span>${paste.Title} - ${new Date(paste.CreatedAt).toLocaleString(undefined, { hour12: false })}</span>
                        <div>
                            <a href="/paste/${paste.Title}" class="btn btn-xs btn-primary">View</a>
                            <button class="btn btn-xs btn-error delete-paste" data-id="${paste.Title}">Delete</button>
                        </div>
                    `;
                    pastesList.appendChild(li);
                });
            }

            showModal(myPastesModal);

            document.querySelectorAll(".delete-paste").forEach((button) => {
                button.addEventListener("click", function () {
                    deletePaste(this.getAttribute("data-id"));
                });
            });
        } catch (error) {
            console.error("Fetch pastes error:", error);
            showToast('Unable to connect to server', 'error');
        }
    });

    async function deletePaste(title) {
        if (confirm(`Are you sure you want to delete: ${title}?`)) {
            try {
                const response = await fetchWithAuth(`/paste/${title}`, { method: 'DELETE' });
                if (response.ok) {
                    const pasteElement = document.querySelector(`[data-id="${title}"]`)?.closest('li');
                    if (pasteElement) pasteElement.remove();
                    showToast('Paste deleted successfully', 'info');
                } else {
                    throw new Error('Failed to delete paste');
                }
            } catch (error) {
                console.error('Delete paste error:', error);
                showToast('Unable to connect to server', 'error');
            }
        }
    }

    document.getElementById('content_modal').addEventListener('click', function(e) {
        if (e.target === this) {
            this.close();
        }
    });

    const modals = document.querySelectorAll('[class*="modal"]');
    modals.forEach(function(modal) {
        modal.addEventListener('click', function(e) {
            if (e.target === this) {
                modal.classList.remove('modal-open');
            }
        });
    });

    createButton.addEventListener('click', async function() {
        const content = document.getElementById('paste-content').value;
        const visibility = document.getElementById('visibility').value;
        const expiration = document.getElementById('expiration').value;
        const password = document.getElementById('encryption').value;

        if (!content.trim()) {
            document.getElementById('content_modal').showModal();
            return;
        }

        let isEncrypted = false;
        let processedContent = content;

        if (password) {
            isEncrypted = true;
            const encoder = new TextEncoder();
            const salt = crypto.getRandomValues(new Uint8Array(16));
            const iv = crypto.getRandomValues(new Uint8Array(12));

            const keyMaterial = await crypto.subtle.importKey(
                'raw', encoder.encode(password), { name: 'PBKDF2' }, false, ['deriveKey']
            );
            const key = await crypto.subtle.deriveKey(
                { name: 'PBKDF2', salt, iterations: 600000, hash: 'SHA-256' },
                keyMaterial, { name: 'AES-GCM', length: 256 }, false,['encrypt']
            );
            const encrypted = await crypto.subtle.encrypt(
                { name: 'AES-GCM', iv }, key, encoder.encode(content)
            );

            // NEW v2 format: [salt (16)][iv (12)][ciphertext]
            const ivAndEncrypted = new Uint8Array(salt.length + iv.length + encrypted.byteLength);
            ivAndEncrypted.set(salt);
            ivAndEncrypted.set(iv, salt.length);
            ivAndEncrypted.set(new Uint8Array(encrypted), salt.length + iv.length);

            processedContent = btoa(String.fromCharCode(...ivAndEncrypted));
        }

        const expirationMap = {
            '10 minutes': '10m', '1 hour': '1h', '1 day': '24h', '3 days': '72h',
            '1 week': '168h', '2 weeks': '336h', '1 month': '720h', 'Never': 'Never',
        };
        const expirationDuration = expirationMap[expiration];
        try {
            const response = await fetch('/paste', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content: processedContent, visibility, expiration: expirationDuration, isEncrypted }),
            });

            if (response.ok) {
                const data = await response.json();
                window.location.href = `/paste/${data.title}`;
            } else {
                const errorContent = await response.json();
                showToast(`${errorContent.message}`, 'error');
            }
        } catch (error) {
            console.error('Create Error:', error);
            showToast('Unable to connect to server', 'error');
        }
        });

        cancelButton.addEventListener('click', function() {
            document.getElementById('paste-content').value = '';
            document.getElementById('visibility').value = 'Public';
            document.getElementById('expiration').value = '1 day';
            document.getElementById('encryption').value = '';
        });
    });

const storedUsername = localStorage.getItem('username');
    if (storedUsername?.trim()) {
        updateAuthUI(storedUsername);
    }

if (document.readyState === 'complete') {
    checkAuthStatus();
} else {
    document.addEventListener('DOMContentLoaded', checkAuthStatus);
}