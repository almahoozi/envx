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

### Key selection
- [ ] Allow selecting the name of the key in the key store
- [ ] Allow exporting/importing keys / salts, etc.

## Priority 2: Shell Integration

### Persistent Configuration
- [ ] Allow setting defaults for different options
- [ ] Allow listing all configs, inluding whether or not the config is a default or override
- [ ] Allow a directory level config override; global config sits in the XDG but if there is
a relevant file in the current dir it merges on top of that
- [ ] Allow setting directory/root specific configs in the global configuration; i.e. profiles;
when we are in the following directory, use the following overrides. These get merged with
globals, before the actual directory config file
- [ ] Allow including other config files (composition)
- [ ] Allow specifying root config file for use when testing (ex: config.test.yml)

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
