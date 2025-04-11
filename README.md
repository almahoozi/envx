# envx - Execute with (encrypted) Environment Variables

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

A limited set of functionality is currently supported:

- Encrypt: `envx encrypt` will print the encrypted version of the `.env`file to
  `stdout`. On first run the encryption key will be created and added to the
  KeyChain. Note this will not overwrite the `.env`file. To do that also pass the
  `-w` flag.
- Run/exec: `envx ./bin/app` will load the `.env`file in the current directory
  into the environment, then execute the binary (not a child process). The env vars
  would be decrypted. If the experimental mechanism above is enabled, `envx` will
  be implicitly prepended to the command; so you can just run `./bin/app` or
  `go run ./cmd/app`.

### Global Flags

- `-n` or `--name`: Name of the env file variant to be used. Appends .{name} to
  the filename. Default is empty. Setting name to local for example will use
  .env.local (if -f is default)
- `-f` or `--file`: Name of the env file to be used. Default is .env. Setting
  the name to .custom will use .custom as the env file. -n is appended to the
  filename if set.
