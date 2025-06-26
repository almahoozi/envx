# Password-Based Encryption Demo

This demonstrates the new password-based encryption feature in envx.

## Key Features

- **Cross-platform**: Works on macOS, Linux, and Windows
- **Secure**: Uses PBKDF2 with 100,000 iterations (OWASP recommended)
- **No password storage**: Passwords are never stored, only prompted when needed
- **Salt-based**: Each account gets a unique salt stored in `~/.config/envx/salts/`

## Usage Examples

### 1. Basic Usage

```bash
# Create a .env file with some secrets
echo "DATABASE_PASSWORD=secret123" > .env
echo "API_KEY=abc123xyz" >> .env

# Encrypt using password-based keystore
envx encrypt --keystore password -w

# The first time, you'll be prompted to create a password:
# Create password for your_username: [hidden input]
# Confirm password for your_username: [hidden input]

# View encrypted file
cat .env
# DATABASE_PASSWORD=ZW52eASuLwa9LpgrE0TAVoDgQls6VaRFHJWcK4kI+llD4SEJml/sJi8g
# API_KEY=ZW52eMwEXhp4xyuBr0d59yj95ufKeoCCuxZGW+e6W32daQqXgiRDYVY=
```

### 2. Adding New Encrypted Variables

```bash
# Add a new variable with secure input (recommended for secrets)
envx add SECRET_TOKEN --keystore password
# Enter password for your_username: [hidden input]
# Enter value for SECRET_TOKEN: [hidden input]

# Or add with value on command line (less secure)
envx add PUBLIC_KEY=ssh-rsa... --keystore password
# Enter password for your_username: [hidden input]
```

### 3. Retrieving Decrypted Values

```bash
# Get all variables
envx get --keystore password
# Enter password for your_username: [hidden input]
# DATABASE_PASSWORD=secret123
# API_KEY=abc123xyz
# SECRET_TOKEN=hidden_value

# Get specific variables
envx get DATABASE_PASSWORD API_KEY --keystore password
# Enter password for your_username: [hidden input]
# DATABASE_PASSWORD=secret123
# API_KEY=abc123xyz

# Get values only
envx getv DATABASE_PASSWORD --keystore password
# Enter password for your_username: [hidden input]
# secret123
```

### 4. Running Programs with Decrypted Environment

```bash
# Run a program with decrypted environment variables
envx run --keystore password ./my-app
# Enter password for your_username: [hidden input]
# [my-app runs with DATABASE_PASSWORD=secret123, etc.]

# Works with any command
envx run --keystore password go run main.go
envx run --keystore password docker-compose up
envx run --keystore password ./scripts/deploy.sh
```

## Cross-Platform Support

### macOS (Production Ready)
- Password keystore: ✅ Full support
- macOS Keychain: ✅ Full support (default)
- Choose based on your needs

### Linux/Windows (Production Ready)
- Password keystore: ✅ Full support (recommended)
- macOS Keychain: ❌ Not available
- Mock keystore: ⚠️ Development only

## Security Details

- **Key Derivation**: PBKDF2 with SHA-256
- **Iterations**: 100,000 (configurable, OWASP minimum)
- **Salt Size**: 256-bit (32 bytes) per account
- **Key Size**: 256-bit (32 bytes) for AES-GCM
- **Salt Storage**: `~/.config/envx/salts/{username}.salt`
- **Password Storage**: None - passwords are never stored

## Migration from Keychain

If you're currently using the macOS Keychain and want to switch to password-based encryption:

```bash
# Export current encrypted values (using keychain)
envx get --json > backup.json

# Switch to password-based encryption
envx encrypt --keystore password -w

# Your .env file is now encrypted with password-based keys
```

## Best Practices

1. **Use secure input for secrets**: `envx add SECRET_KEY` (prompts securely)
2. **Avoid command-line values for secrets**: Don't use `envx add SECRET=value`
3. **Use strong passwords**: The security depends on your password strength
4. **Backup your salt files**: Store `~/.config/envx/salts/` securely for recovery
5. **Use different passwords per project**: Each directory can have different encryption

## Troubleshooting

### "Failed to read salt file" error
- This happens on first use - the salt will be created automatically
- If salt file is corrupted, delete it and create a new password

### "Passwords do not match" error
- Ensure you type the same password when creating
- Use a password manager to generate and store strong passwords

### "Inappropriate ioctl for device" error
- This happens when trying to pipe passwords or run in non-interactive mode
- Run commands in an interactive terminal that supports password input

## Implementation Details

The password-based keystore:
1. Generates a random 256-bit salt per account
2. Stores salt in `~/.config/envx/salts/{account}.salt`
3. Derives encryption key using PBKDF2(password, salt, 100000, SHA-256)
4. Uses derived key with AES-GCM for encryption/decryption
5. Never stores passwords - always prompts when needed 