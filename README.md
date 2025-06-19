# envx - Execute with (encrypted) Environment Variables

[![CI](https://github.com/almahoozi/envx/actions/workflows/ci.yml/badge.svg)](https://github.com/almahoozi/envx/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/almahoozi/envx)](https://github.com/almahoozi/envx/blob/main/go.mod)
[![License](https://img.shields.io/github/license/almahoozi/envx)](https://github.com/almahoozi/envx/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/almahoozi/envx)](https://github.com/almahoozi/envx/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/almahoozi/envx)](https://goreportcard.com/report/github.com/almahoozi/envx)
[![Security](https://img.shields.io/badge/security-passing-brightgreen)](https://github.com/almahoozi/envx/actions/workflows/ci.yml)
[![Platform](https://img.shields.io/badge/platform-macOS%20(production)%20%7C%20Linux/Windows%20(testing)-blue)](#platform-support)

```bash
go install github.com/almahoozi/envx@latest
```

Assuming `$GOPATH/bin` is in your `$PATH`, otherwise go do that.

A nifty mechanism that I'm experimenting with is to automatically prepend the
`go run` and `./*` commands with `envx` for convenience. This is only done if
there is a `.env` file in the current directory, and the current directory is
not `envx` itself.

```bash
function rewrite_buffer() {
    if [[ -f .env && "${PWD##*/}" != "envx" && ( "$BUFFER" == go\ run* || "$BUFFER" == ./* ) ]]; then
      BUFFER="envx $BUFFER"
    fi
}

zle -N zle-line-finish rewrite_buffer
```

## Commands

envx provides several commands for managing encrypted environment variables:

### `run` - Execute Programs with Decrypted Environment (Default Command)
```bash
envx run ./bin/app
envx ./bin/app        # equivalent - run is the default command
envx run go run main.go
envx go run main.go   # equivalent
```
Loads the `.env` file, decrypts all values, sets them as environment variables, and executes the specified program. 

**The `run` subcommand is optional** - if no recognized subcommand is provided, `envx` defaults to the `run` behavior. However, explicitly specifying `run` is useful for:
- Disambiguation when your executable name conflicts with an `envx` command
- Clarity in scripts and documentation
- Avoiding confusion with command-line arguments

If the experimental mechanism above is enabled, `envx` will be implicitly prepended to commands.

### `encrypt` - Encrypt Environment Variables
```bash
envx encrypt                    # encrypt all variables, print to stdout
envx encrypt KEY1 KEY2          # encrypt specific variables only
envx encrypt -w                 # encrypt and overwrite the .env file
envx encrypt --json             # output in JSON format
```
Encrypts unencrypted variables in the `.env` file. By default prints to stdout; use `-w` to overwrite the file.

### `decrypt` - Decrypt Environment Variables
```bash
envx decrypt                    # decrypt all variables, print to stdout
envx decrypt KEY1 KEY2          # decrypt specific variables only
envx decrypt -w                 # decrypt and overwrite the .env file
envx decrypt --json             # output in JSON format
```
Decrypts encrypted variables in the `.env` file. By default prints to stdout; use `-w` to overwrite the file.

### `add` - Add New Encrypted Variables
```bash
envx add KEY1=value1 KEY2=value2    # add new variables (fails if exists)
envx add KEY1 KEY2                  # prompt securely for values (recommended for secrets)
envx add KEY=value -p               # print result instead of writing
envx add KEY=value --json           # output in JSON format
```
Encrypts and adds new variables to the `.env` file. Fails if the variable already exists (use `set` to overwrite).

**Secure Input**: You can specify just the key names (without `=value`) and envx will prompt you to enter the values securely without echoing to the terminal or storing them in shell history. This is the recommended approach for sensitive values like passwords and API keys.

### `set` - Set/Update Encrypted Variables
```bash
envx set KEY1=value1 KEY2=value2    # set variables (overwrites if exists)
envx set KEY1 KEY2                  # prompt securely for values (recommended for secrets)
envx set KEY=value -p               # print result instead of writing
envx set KEY=value --json           # output in JSON format
```
Encrypts and sets variables in the `.env` file. Overwrites existing values (use `add` to prevent overwriting).

**Secure Input**: Like `add`, you can specify just key names and envx will prompt securely for values without exposing them in terminal history.

### `get` - Retrieve Decrypted Variables
```bash
envx get                        # get all variables
envx get KEY1 KEY2              # get specific variables
envx get --json                 # output in JSON format
envx get -v                     # values only (no keys)
```
Retrieves and decrypts variables from the `.env` file.

### `getv` - Get Values with Custom Separator
```bash
envx getv                       # get all values (newline separated)
envx getv KEY1 KEY2             # get specific values
envx getv -s ","                # use comma separator
```
Retrieves only the values (not keys) of decrypted variables with a customizable separator.

### `man` - Show Manual
```bash
envx man
```
Displays the embedded manual page.

## Global Flags

All commands support these flags:

- `-n` or `--name`: Name of the env file variant to be used. Appends `.{name}` to the filename. Default is empty. Setting name to `local` for example will use `.env.local` (if `-f` is default)
- `-f` or `--file`: Name of the env file to be used. Default is `.env`. Setting the name to `.custom` will use `.custom` as the env file. `-n` is appended to the filename if set.

## Format Options

Commands that output data support format options:

- `-F` or `--fmt`: Specify format explicitly (`env`, `json`, `yaml`)
- `-j` or `--json`: Output in JSON format
- `-y` or `--yml` or `--yaml`: Output in YAML format (note: YAML is not yet fully implemented)

## Write Options

Commands that modify files support:

- `-w` or `--write`: Write changes to the file instead of printing to stdout
- `-p` or `--print`: Print output instead of writing to file (for `add`/`set` commands)

## Platform Support

### Production Support
- **macOS**: Full production support with secure keychain integration
- All encryption keys are stored in the macOS Keychain for maximum security

### Testing/Development Support  
- **Linux/Windows**: Functional for testing and development
- Uses in-memory mock keystore (keys are not persisted between sessions)
- Suitable for CI/CD pipelines and development environments
- Not recommended for production use due to non-persistent key storage

## Security Features

### Secure Value Input
For maximum security when adding or updating secrets:
- **Use key-only format**: `envx add SECRET_KEY` instead of `envx add SECRET_KEY=value`
- envx will prompt you to enter the value securely without echoing to the terminal
- Values won't appear in your shell history or process list
- Recommended for passwords, API keys, tokens, and other sensitive data

### Key Management
envx automatically manages encryption keys using the macOS Keychain:
- Keys are generated automatically on first use
- Keys are stored securely in the system keychain
- Keys are retrieved automatically for encryption/decryption operations

## Examples

```bash
# Run Go programs with decrypted environment
envx run go run main.go
envx run ./my-server

# Run other programs with decrypted environment  
envx run ./scripts/deploy.sh
envx run docker-compose up

# Add secrets securely (prompts for values without exposing them)
envx add DATABASE_PASSWORD API_KEY

# Add a new secret with value on command line (less secure)
envx add DATABASE_PASSWORD=secret123

# Update existing secrets securely
envx set API_KEY JWT_SECRET

# Update an existing value with command line (less secure)
envx set API_KEY=new-api-key-value

# View all decrypted values
envx get

# Get specific values as JSON
envx get DATABASE_URL API_KEY --json

# Encrypt all plain text values in .env
envx encrypt -w

# Work with different env files
envx -f .env.production get
envx -n staging add NEW_VAR
```
