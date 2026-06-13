const CryptoHelper = {
    async generateKey() {
        return await window.crypto.subtle.generateKey(
            { name: "AES-GCM", length: 256 },
            true,
            ["encrypt", "decrypt"]
        );
    },

    async exportKey(key) {
        const exported = await window.crypto.subtle.exportKey("raw", key);
        return this.arrayBufferToBase64(exported);
    },

    async importKey(base64Key) {
        const rawKey = this.base64ToArrayBuffer(base64Key);
        return await window.crypto.subtle.importKey(
            "raw", rawKey, "AES-GCM", true, ["encrypt", "decrypt"]
        );
    },

    async derivePasscodeKey(passcode, saltBase64) {
        const salt = this.base64ToArrayBuffer(saltBase64);
        const result = await argon2.hash({
            pass: passcode,
            salt: new Uint8Array(salt),
            time: 2,
            mem: 16384,
            hashLen: 32,
            parallelism: 1,
            type: argon2.ArgonType.Argon2id
        });
        
        return await window.crypto.subtle.importKey(
            "raw", result.hash, "AES-GCM", true, ["encrypt", "decrypt"]
        );
    },

    generateSalt() {
        const salt = window.crypto.getRandomValues(new Uint8Array(16));
        return this.arrayBufferToBase64(salt);
    },

    async encrypt(plaintext, key) {
        if (typeof key === 'string') key = await this.importKey(key);
        const encoder = new TextEncoder();
        const data = encoder.encode(plaintext);
        const iv = window.crypto.getRandomValues(new Uint8Array(12));

        const ciphertext = await window.crypto.subtle.encrypt(
            { name: "AES-GCM", iv: iv },
            key, data
        );

        const combined = new Uint8Array(iv.length + ciphertext.byteLength);
        combined.set(iv);
        combined.set(new Uint8Array(ciphertext), iv.length);

        return this.arrayBufferToBase64(combined);
    },

    async decrypt(base64Ciphertext, key) {
        if (typeof key === 'string') key = await this.importKey(key);
        const combined = this.base64ToArrayBuffer(base64Ciphertext);
        const iv = combined.slice(0, 12);
        const ciphertext = combined.slice(12);

        const decrypted = await window.crypto.subtle.decrypt(
            { name: "AES-GCM", iv: iv },
            key, ciphertext
        );

        const decoder = new TextDecoder();
        return decoder.decode(decrypted);
    },

    arrayBufferToBase64(buffer) {
        let binary = '';
        const bytes = new Uint8Array(buffer);
        for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i]);
        }
        return window.btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
    },

    base64ToArrayBuffer(base64) {
        base64 = base64.replace(/-/g, '+').replace(/_/g, '/');
        while (base64.length % 4) base64 += '=';
        const binary_string = window.atob(base64);
        const bytes = new Uint8Array(binary_string.length);
        for (let i = 0; i < binary_string.length; i++) {
            bytes[i] = binary_string.charCodeAt(i);
        }
        return bytes.buffer;
    }
};

document.addEventListener('DOMContentLoaded', () => {
    const app = {
        views: document.querySelectorAll('.view'),
        createBtn: document.getElementById('create-btn'),
        secretInput: document.getElementById('secret-input'),
        expirySelect: document.getElementById('expiry-select'),
        viewLimitSelect: document.getElementById('view-limit'),
        passcodeEnable: document.getElementById('passcode-enable'),
        passcodeInput: document.getElementById('passcode-input'),
        shareUrlInput: document.getElementById('share-url'),
        copyBtn: document.getElementById('copy-btn'),
        newSecretBtn: document.getElementById('new-secret-btn'),
        viewBtn: document.getElementById('view-btn'),
        viewPasscode: document.getElementById('view-passcode'),
        secretOutput: document.getElementById('secret-output'),
        copySecretBtn: document.getElementById('copy-secret-btn'),
        viewNewBtn: document.getElementById('view-new-btn'),
        retryBtn: document.getElementById('retry-btn'),

        meta: null,

        init() {
            this.bindEvents();
            this.route();
        },

        bindEvents() {
            this.createBtn.addEventListener('click', () => this.createSecret());
            this.copyBtn.addEventListener('click', () => this.copyToClipboard(this.shareUrlInput));
            this.copySecretBtn.addEventListener('click', () => this.copyToClipboard(this.secretOutput));
            this.newSecretBtn.addEventListener('click', () => this.resetApp());
            this.viewNewBtn.addEventListener('click', () => this.resetApp());
            this.viewBtn.addEventListener('click', () => this.fetchSecret());
            this.retryBtn.addEventListener('click', () => this.resetApp());

            this.passcodeEnable.addEventListener('change', (e) => {
                this.passcodeInput.classList.toggle('hidden', !e.target.checked);
            });

            window.addEventListener('hashchange', () => {
                this.route();
            });
        },

        route() {
            const path = window.location.pathname;
            const hash = window.location.hash;

            if (path.startsWith('/s/')) {
                const parts = path.split('/');
                const id = parts[2];
                
                if (!id) {
                    return this.showView('create-view');
                }

                this.currentSecretId = id;
                this.currentKey = hash.substring(1);
                this.meta = null;
                this.showView('decrypt-view');
            } else {
                this.showView('create-view');
            }
        },

        showView(id) {
            this.views.forEach(v => v.classList.add('hidden'));
            const el = document.getElementById(id);
            if (el) el.classList.remove('hidden');
        },

        async createSecret() {
            const plaintext = this.secretInput.value;
            if (!plaintext) return alert('Please enter a secret.');

            try {
                this.createBtn.disabled = true;
                this.createBtn.textContent = 'Creating...';

                const key = await CryptoHelper.generateKey();
                const exportedKey = await CryptoHelper.exportKey(key);
                let ciphertext = await CryptoHelper.encrypt(plaintext, key);
                let salt = "";

                if (this.passcodeEnable.checked) {
                    const passcode = this.passcodeInput.value;
                    if (!passcode) throw new Error('Passcode is required');
                    salt = CryptoHelper.generateSalt();
                    const passcodeKey = await CryptoHelper.derivePasscodeKey(passcode, salt);
                    ciphertext = await CryptoHelper.encrypt(ciphertext, passcodeKey);
                }

                const response = await fetch('/api/v1/secrets', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        ciphertext: ciphertext,
                        expiration_minutes: parseInt(this.expirySelect.value),
                        view_limit: parseInt(this.viewLimitSelect.value),
                        passcode_enabled: this.passcodeEnable.checked,
                        salt: salt
                    })
                });

                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(`Server Error (${response.status}): ${errorText}`);
                }
                const data = await response.json();
                
                const shareUrl = `${window.location.origin}/s/${data.id}#${exportedKey}`;
                this.shareUrlInput.value = shareUrl;
                this.showView('success-view');

            } catch (err) {
                this.showError('Creation Failed', err.message);
            } finally {
                this.createBtn.disabled = false;
                this.createBtn.textContent = 'Create Secure Link';
            }
        },

        async fetchSecret() {
            try {
                this.viewBtn.disabled = true;
                this.viewBtn.textContent = 'Decrypting...';

                if (!this.meta) {
                    const metaRes = await fetch(`/api/v1/meta/${this.currentSecretId}`);
                    if (!metaRes.ok) {
                        if (metaRes.status === 404) return this.showError('Not Found', 'Secret expired or destroyed.');
                        const errorText = await metaRes.text();
                        throw new Error(`Meta Error (${metaRes.status}): ${errorText}`);
                    }
                    this.meta = await metaRes.json();
                }

                let passcode = "";
                if (this.meta.passcode_enabled) {
                    passcode = this.viewPasscode.value;
                    if (!passcode) {
                        document.getElementById('passcode-section').classList.remove('hidden');
                        document.getElementById('decrypt-msg').textContent = 'This secret is passcode protected.';
                        this.viewBtn.disabled = false;
                        this.viewBtn.textContent = 'View Secret';
                        return;
                    }
                }

                const response = await fetch(`/api/v1/secrets/${this.currentSecretId}`);
                if (!response.ok) {
                    if (response.status === 404) return this.showError('Not Found', 'Secret expired or destroyed.');
                    const errorText = await response.text();
                    throw new Error(`Retrieval Error (${response.status}): ${errorText}`);
                }

                const data = await response.json();
                let ciphertext = data.ciphertext;

                if (data.passcode_enabled) {
                    const passcodeKey = await CryptoHelper.derivePasscodeKey(passcode, data.salt);
                    ciphertext = await CryptoHelper.decrypt(ciphertext, passcodeKey);
                }

                const key = await CryptoHelper.importKey(this.currentKey);
                const plaintext = await CryptoHelper.decrypt(ciphertext, key);

                this.secretOutput.value = plaintext;
                this.showView('view-view');

            } catch (err) {
                let userMsg = err.message;
                const isCryptoErr = err.name === 'OperationError' || err.message.includes('AES-GCM') || err.message.includes('decrypt');
                const isPasscodeErr = err.message.includes('passcode') || (isCryptoErr && this.meta && this.meta.passcode_enabled);

                if (isPasscodeErr) {
                    userMsg = 'Incorrect passcode or invalid link. Please try again.';
                } else if (isCryptoErr) {
                    userMsg = 'Decryption failed. The link may be corrupted or invalid.';
                }

                this.showError('Decryption Failed', userMsg);
                
                if (isPasscodeErr && this.meta && this.meta.passcode_enabled) {
                    setTimeout(() => {
                        this.showView('decrypt-view');
                        this.viewBtn.disabled = false;
                        this.viewBtn.textContent = 'View Secret';
                    }, 2000);
                }
            } finally {
                if (document.getElementById('view-view').classList.contains('hidden')) {
                    this.viewBtn.disabled = false;
                    this.viewBtn.textContent = 'View Secret';
                }
            }
        },

        showError(title, msg) {
            document.getElementById('error-title').textContent = title;
            document.getElementById('error-msg').textContent = msg;
            this.showView('error-view');
        },

        copyToClipboard(inputEl) {
            inputEl.select();
            document.execCommand('copy');
            const btn = inputEl.nextElementSibling;
            if (btn && btn.tagName === 'BUTTON') {
                const originalText = btn.textContent;
                btn.textContent = 'Copied!';
                setTimeout(() => btn.textContent = originalText, 2000);
            }
        },

        resetApp() {
            window.location.href = '/';
        }
    };

    app.init();
});
