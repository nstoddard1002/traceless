# traceless.pw

## Minimum Viable Product (MVP) Specification

### Version

0.1

### Status

Draft

### Purpose

traceless.pw is a secure, privacy-focused secret sharing service that allows users to share passwords, notes, API keys, recovery codes, and other sensitive information through self-destructing links.

The primary design goal is to make secure sharing simple while ensuring that unauthorized parties—including the service operator—cannot access shared content.

---

# Problem Statement

Sensitive information is frequently shared through insecure channels such as:

* Email
* SMS
* Slack
* Teams
* Discord
* Ticketing systems

These systems often retain messages indefinitely and may expose information to unauthorized individuals.

traceless.pw provides a temporary, encrypted mechanism for sharing secrets that automatically expire and can optionally self-destruct after viewing.

---

# MVP Goals

The MVP must:

* Allow users to create secure notes without creating an account.
* Encrypt all content before it reaches the server.
* Support one-time viewing.
* Support expiration times.
* Support optional passcode protection.
* Permanently delete expired or viewed secrets.
* Provide a simple user experience requiring minimal technical knowledge.

---

# Non-Goals (MVP)

The MVP will NOT include:

* User accounts
* Teams
* Audit logs
* File attachments
* Encrypted file storage
* Mobile applications
* Custom domains
* Enterprise features
* Cryptocurrency payments
* Geofencing
* Decoy notes
* Multi-party authorization

These may be considered for future releases.

---

# User Stories

## Secret Creator

As a user,

I want to create a secure note and receive a shareable link,

so that I can safely send sensitive information.

---

## Secret Recipient

As a recipient,

I want to access the note using a secure link,

so that I can retrieve the information I need.

---

## Security-Conscious User

As a security-conscious user,

I want the note to self-destruct after viewing,

so that it cannot be accessed later.

---

## Privacy-Conscious User

As a privacy-conscious user,

I want the server to be unable to read my secret,

so that a server compromise does not expose my information.

---

# Functional Requirements

## Create Secret

The user enters:

* Secret text
* Expiration period
* View limit
* Optional passcode

The system returns:

* Shareable URL

Example:

https://traceless.pw/s/abc123#KEY

---

## Secret Types

MVP supports:

* Plain text
* Passwords
* API keys
* Recovery codes
* Notes

Maximum size:

16 KB

---

## Expiration Options

Available presets:

* 10 minutes
* 1 hour
* 24 hours
* 7 days

Default:

24 hours

---

## View Limits

Available options:

* One view
* Three views
* Five views

Default:

One view

---

## Passcode Protection

Optional.

If enabled:

Recipient must provide the passcode before the secret is revealed.

Passcodes are never stored in plaintext.

---

## Burn After Read

When a secret reaches its view limit:

* Secret becomes unavailable
* Database record is deleted

User receives:

"This secret is no longer available."

---

## Expiration Cleanup

Background worker removes:

* Expired secrets
* Destroyed secrets

Cleanup interval:

Every 5 minutes

---

# Security Requirements

## End-to-End Encryption

Secrets must be encrypted in the browser.

Server must never receive plaintext content.

---

## Encryption Algorithm

AES-256-GCM

---

## Key Generation

Browser generates:

* Random 256-bit encryption key

Key remains client-side.

---

## URL Fragment Storage

Encryption key stored after URL fragment:

https://traceless.pw/s/abc123#KEY

The fragment is never transmitted to the server.

---

## Passcode Protection

Passcode derives a secondary key.

Recommended algorithm:

Argon2id

---

## Transport Security

All communication requires:

* HTTPS
* TLS 1.3

HTTP redirects to HTTPS.

---

## Rate Limiting

Apply limits:

* Per IP
* Per secret identifier

Suggested:

10 requests per minute

---

## Content Security Policy

Strict CSP enabled.

No inline scripts.

---

## Logging Policy

Do not log:

* Secret contents
* URL fragments
* Passcodes

Minimal operational logging only.

---

# User Interface

## Home Page

Elements:

* Secret text area
* Expiration selector
* View limit selector
* Optional passcode field
* Create Secret button

---

## Success Page

Displays:

* Secure URL
* Copy button

Warning:

"Anyone with this link may access the secret."

---

## Secret Access Page

Displays:

* Passcode field (if required)
* View Secret button

---

## Secret View Page

Displays:

* Secret content
* Copy button

Warning:

"This secret may self-destruct after viewing."

---

## Destroyed Secret Page

Displays:

"This secret is no longer available."

---

# Backend Architecture

## API Endpoints

### POST /api/v1/secrets

Creates a secret.

Returns:

* Secret ID

---

### GET /api/v1/secrets/{id}

Returns encrypted payload.

---

### DELETE /api/v1/secrets/{id}

Internal use only.

Used by cleanup worker.

---

# Database Schema

Table: secrets

Fields:

id
varchar(32)

ciphertext
text

created_at
timestamp

expires_at
timestamp

remaining_views
integer

passcode_enabled
boolean

salt
varchar(128)

---

# Technology Stack

Frontend:

* HTML
* CSS
* JavaScript
* WebCrypto API

Backend:

* Go

Database:

* PostgreSQL

Deployment:

* Docker
* Linux VPS

Reverse Proxy:

* Nginx

---

# Success Criteria

The MVP is considered successful when:

1. Users can create secure notes.
2. Notes are encrypted before upload.
3. Notes can be shared via URL.
4. One-time view functionality works reliably.
5. Expired notes are automatically deleted.
6. No plaintext secret data is stored server-side.
7. Average secret creation time is under 5 seconds.

---

# Future Enhancements

Version 2 candidates:

* User accounts
* File sharing
* QR code sharing
* Enterprise deployments
* Self-hosted edition
* API access
* Team workspaces
* Audit events
* Secret revocation
* Custom expiration policies
* Anonymous Tor endpoint

---

# Product Tagline

Share secrets. Leave nothing behind.
