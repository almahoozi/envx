# envx - Execute with (encrypted) Environment Variables

```bash
go install github.com/almahoozi/envx@latest
```

Assuming `$GOPATH/bin` is in your `$PATH`, otherwise go do that.

A nifty mechanism that I'm experimenting with is to automatically prepend the
`go run` and `./*` commands with `envx` for convenience.

```bash
function rewrite_buffer() {
    if [[ "$BUFFER" == go\ run* || "$BUFFER" == ./* ]]; then
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
