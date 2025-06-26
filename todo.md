# envx TODO List

## Priority 1: Advanced Encryption Options

### YubiKey PIV Support
- [ ] Implement YubiKey PIV integration for hardware-based cryptography
- [ ] Add PIV key generation and management
- [ ] Support for PIV slots and certificates
- [ ] Hardware-based signing and encryption operations

### macOS Secure Enclave Support
- [ ] Implement Secure Enclave integration for macOS
- [ ] Use Secure Enclave for key generation and storage
- [ ] Hardware-based cryptographic operations in Secure Enclave
- [ ] Fallback to Keychain when Secure Enclave unavailable

### Password-Based Encryption
- [x] Implement password-based key derivation (PBKDF2/Argon2)
- [x] Password prompt for encryption/decryption operations
- [x] No password storage - request on each operation
- [x] Cross-platform support (works on Linux/Windows/macOS)
- [x] Configurable key derivation parameters

## Priority 2: Shell Integration

### Auto-completion
- [ ] Command and option completion for bash/zsh/fish
- [ ] Environment variable key completion from .env files
- [ ] Dynamic completion based on current directory
- [ ] Installation scripts for shell completion

### Shell Alias Resolution
- [ ] Resolve shell aliases before execution
- [ ] Support for bash/zsh alias detection
- [ ] Integration with shell built-ins and functions

## Lower Hanging Fruit

### Security Improvements
- [ ] Randomized magic bytes (replace fixed "envx" prefix)
- [ ] Implement secure random prefix generation
- [ ] Update encryption/decryption to handle variable prefixes

### Key Import/Export
- [ ] `export-key` command - export key in plaintext or password-encrypted
- [ ] `import-key` command - import key from file
- [ ] Support for encrypted key export with password protection
- [ ] Cross-machine key migration support

## Implementation Notes

- Start with password-based encryption as it provides immediate cross-platform benefits
- YubiKey PIV requires hardware dependencies and testing
- Secure Enclave is macOS-specific but provides excellent security
- Shell completion can be implemented incrementally per shell
