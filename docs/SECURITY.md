# Security Model

traceless.pw is built on a **Zero-Knowledge** architecture. This means the server never sees the plaintext content of your secrets or the keys required to decrypt them.

## Encryption Flow

### 1. Key Generation
When you create a secret, your browser generates a random 256-bit AES key using the `crypto.subtle.generateKey` API.

### 2. Encryption
The plaintext is encrypted using **AES-256-GCM**.
- A random 12-byte IV (Initialization Vector) is generated for each encryption.
- The IV is prepended to the ciphertext.

### 3. Key Storage (Client-Side)
The encryption key is exported as a Base64 string and appended to the shareable URL as a **URL Fragment** (the part after the `#`).
- **Crucial Detail:** URL fragments are never sent to the server by browsers during HTTP requests. This ensures the key remains strictly client-side.

### 4. Passcode Protection (Optional)
If a passcode is enabled, the secret undergoes **Double Encryption**:
1. The plaintext is encrypted with the `URL_Key` -> `Encrypted_Level_1`.
2. A secondary key (`Passcode_Key`) is derived from the user's passcode and a random salt using **Argon2id**.
3. `Encrypted_Level_1` is encrypted with `Passcode_Key` -> `Final_Ciphertext`.
4. The `Final_Ciphertext` and `Salt` are sent to the server.

To view the secret, the recipient needs BOTH the `URL_Key` (from the link) and the correct `Passcode`.

## Data Retention

Secrets are stored in a PostgreSQL database and are purged in two ways:
1. **On-Access:** When the view limit is reached, the backend immediately deletes the record.
2. **Background Worker:** A background routine runs every 5 minutes to delete any secrets that have passed their expiration timestamp.

## Limitations

- **Browser Security:** While the server is zero-knowledge, the security of the encryption depends on the integrity of the JavaScript served to your browser. Use a trusted instance of traceless.pw.
- **Link Exposure:** Anyone with the full URL (including the fragment) can access the secret unless a passcode is enabled.
