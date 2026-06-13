CREATE TABLE IF NOT EXISTS secrets (
    id VARCHAR(32) PRIMARY KEY,
    ciphertext TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    remaining_views INTEGER NOT NULL,
    passcode_enabled BOOLEAN DEFAULT FALSE,
    salt VARCHAR(128)
);

CREATE INDEX idx_secrets_expires_at ON secrets (expires_at);
