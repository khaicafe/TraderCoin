# API Credentials Encryption

## Overview

All API keys and secrets are encrypted using AES-256-GCM before being stored in the database.

## Configuration

### Environment Variable

Set the encryption key in your `.env` file:

```bash
ENCRYPTION_KEY=your-32-byte-encryption-key-here-12345
```

**Important:** The key must be exactly 32 bytes for AES-256.

## Implementation

### Encryption Functions

- `EncryptString(plaintext string)` - Encrypts a string using AES-256-GCM
- `DecryptString(ciphertext string)` - Decrypts an encrypted string
- `GetDecryptedAPICredentials(config *TradingConfig)` - Helper to get decrypted credentials

### When Encryption Happens

1. **Creating Bot Config:**

   - API Key and API Secret are encrypted before saving to database
   - Empty strings are not encrypted

2. **Updating Bot Config:**

   - New API Key/Secret values are encrypted before update
   - Existing encrypted values remain unchanged if not provided

3. **Using API Credentials:**
   - Decrypt when needed to call exchange APIs
   - Use `GetDecryptedAPICredentials()` helper function

## Security Best Practices

1. ✅ **Never log decrypted keys**
2. ✅ **Use environment variables for encryption key**
3. ✅ **Rotate encryption key periodically**
4. ✅ **API credentials are never exposed in JSON responses** (`json:"-"` tag)
5. ✅ **Use HTTPS in production**

## Example Usage

```go
// When creating a bot config
encryptedKey, err := EncryptString(apiKey)
if err != nil {
    return err
}
config.APIKey = encryptedKey

// When calling exchange API
apiKey, apiSecret, err := GetDecryptedAPICredentials(&config)
if err != nil {
    return err
}
// Use apiKey and apiSecret to call exchange
```

## Migration

If you have existing unencrypted data:

1. Create a migration script to encrypt existing API keys
2. Use the same encryption functions
3. Update all records in the database

## Testing

The encryption key is set during application initialization in `main.go`:

```go
controllers.InitEncryptionKey(cfg.EncryptionKey)
```
