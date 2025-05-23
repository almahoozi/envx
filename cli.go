package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	flag "github.com/spf13/pflag"
)

type command[T any] struct {
	flags *flag.FlagSet
	fn    func(context.Context, T, ...string) error
	val   T
}

func (c *command[T]) execute(ctx context.Context, args ...string) error {
	if err := c.flags.Parse(args); err != nil {
		return err
	}
	return c.fn(ctx, c.val, c.flags.Args()...)
}

type Format string

const (
	FormatEnv  Format = "env"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

type fmtOpts struct {
	format string
	json   bool
	yaml   bool
	yml    bool
}

func NewFmtOpts(flags *flag.FlagSet) *fmtOpts {
	opts := new(fmtOpts)
	flags.StringVarP(&opts.format, "fmt", "F", "", "Format of the output. Supported formats: env, json, yaml (default \"env\")")
	// TODO: Consider whether we really want the shorthands
	flags.BoolVarP(&opts.json, "json", "j", false, "Format the output to JSON")
	flags.BoolVar(&opts.yaml, "yaml", false, "Format the output to YAML") // TODO: Change to an alias
	flags.BoolVarP(&opts.yml, "yml", "y", false, "Format the output to YAML")
	flags.SortFlags = false
	return opts
}

func (opts *fmtOpts) Format() (Format, error) {
	switch Format(opts.format) {
	case FormatEnv, FormatJSON, FormatYAML:
		if opts.json || opts.yaml || opts.yml {
			fmt.Println(opts.json, opts.yaml, opts.yml, opts.format)
			return "", fmt.Errorf("cannot use both format and json/yaml/yml flags")
		}
		return Format(opts.format), nil
	}

	if opts.json {
		if opts.yaml || opts.yml {
			return "", fmt.Errorf("cannot use both json and yaml/yml flags")
		}
		return FormatJSON, nil
	}

	if opts.yaml || opts.yml {
		if opts.yaml && opts.yml {
			return "", fmt.Errorf("cannot use both yaml and yml flags")
		}

		return FormatYAML, nil
	}

	return FormatEnv, nil
}

type encryptOpts struct {
	Name    string
	File    string
	FmtOpts *fmtOpts
	Write   bool
}

type decryptOpts struct {
	Name    string
	File    string
	FmtOpts *fmtOpts
	Write   bool
}

type runOpts struct {
	Name string
	File string
	Args []string
}

type executor interface {
	execute(ctx context.Context, args ...string) error
}

func start() error {
	cmds := make(map[string]executor)

	manCmd := new(command[struct{}])
	manCmd.flags = flag.NewFlagSet("man", flag.ExitOnError)
	manCmd.fn = func(context.Context, struct{}, ...string) error {
		fmt.Println(man)
		return nil
	}
	cmds[manCmd.flags.Name()] = manCmd

	runCmd := new(command[runOpts])
	runCmd.flags = flag.NewFlagSet("run", flag.ExitOnError)
	runCmd.flags.StringVarP(&runCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	runCmd.flags.StringVarP(&runCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	runCmd.fn = run
	cmds[runCmd.flags.Name()] = runCmd

	encCmd := new(command[encryptOpts])
	encCmd.flags = flag.NewFlagSet("encrypt", flag.ExitOnError)
	encCmd.flags.StringVarP(&encCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	encCmd.flags.StringVarP(&encCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	encCmd.val.FmtOpts = NewFmtOpts(encCmd.flags)
	encCmd.flags.BoolVarP(&encCmd.val.Write, "write", "w", false, "Overwrites the file with encrypted values.")
	encCmd.fn = encryptCmd
	cmds[encCmd.flags.Name()] = encCmd

	decCmd := new(command[decryptOpts])
	decCmd.flags = flag.NewFlagSet("decrypt", flag.ExitOnError)
	decCmd.flags.StringVarP(&decCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	decCmd.flags.StringVarP(&decCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	decCmd.val.FmtOpts = NewFmtOpts(decCmd.flags)
	decCmd.flags.BoolVarP(&decCmd.val.Write, "write", "w", false, "Overwrites the file with decrypted values.")
	decCmd.fn = decryptCmd
	cmds[decCmd.flags.Name()] = decCmd

	cmds[""] = runCmd

	if len(os.Args) >= 2 {
		if cmd, ok := cmds[os.Args[1]]; ok {
			return cmd.execute(context.Background(), os.Args[2:]...)
		}
	}

	if cmd, ok := cmds[""]; ok {
		return cmd.execute(context.Background(), os.Args[1:]...)
	}

	return fmt.Errorf("missing command")
}

func encryptCmd(ctx context.Context, opts encryptOpts, args ...string) error {
	format, err := opts.FmtOpts.Format()
	if err != nil {
		return fmt.Errorf("error parsing format: %w", err)
	}
	if format == FormatYAML {
		return fmt.Errorf("unsupported format: %s", format)
	}

	file := opts.File
	if opts.Name != "" {
		file = fmt.Sprintf("%s.%s", file, opts.Name)
	}

	key, err := loadKey()
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	// Load .env file if exists
	vars, err := loadEnv(file, key)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	argMap := make(map[string]bool, len(args))
	for _, arg := range args {
		argMap[arg] = true
	}

	for i, v := range vars {
		// If it isn't already encrypted, encrypt it
		if len(args) == 0 || argMap[v.key] {
			ciphertext, err := encrypt(v.value, key)
			if err != nil {
				return fmt.Errorf("error encrypting value: %w", err)
			}
			vars[i].value = ciphertext
		}
	}

	var sb strings.Builder
	jsonVars := make([]string, 0, len(vars))
	for _, v := range vars {
		switch format {
		case FormatJSON:
			// WARN: Obviously this is not the safest as we don't check for duplicates
			jsonVars = append(jsonVars, fmt.Sprintf("%q:%q", v.key, v.value))
		default:
			sb.WriteString(fmt.Sprintf("%s=%s\n", v.key, v.value))
		}
	}

	if format == FormatJSON {
		sb.WriteString("{")
		sb.WriteString(strings.Join(jsonVars, ","))
		sb.WriteString("}")
	}

	output := sb.String()
	if !opts.Write {
		fmt.Println(output)
		return nil
	}

	if err := os.WriteFile(file, []byte(output), 0o644); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
}

func decryptCmd(ctx context.Context, opts decryptOpts, args ...string) error {
	format, err := opts.FmtOpts.Format()
	if err != nil {
		return fmt.Errorf("error parsing format: %w", err)
	}
	if format == FormatYAML {
		return fmt.Errorf("unsupported format: %s", format)
	}

	file := opts.File
	if opts.Name != "" {
		file = fmt.Sprintf("%s.%s", file, opts.Name)
	}

	key, err := loadKey()
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	// Load .env file if exists
	vars, err := loadEnv(file, key)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	argMap := make(map[string]bool, len(args))
	for _, arg := range args {
		argMap[arg] = true
	}

	for i, v := range vars {
		if len(args) == 0 || argMap[v.key] {
			plaintext, err := decrypt(v.value, key)
			if err != nil {
				return fmt.Errorf("error decrypting value: %w", err)
			}
			vars[i].value = plaintext
		}
	}

	var sb strings.Builder
	jsonVars := make([]string, 0, len(vars))
	for _, v := range vars {
		switch format {
		case FormatJSON:
			// WARN: Obviously this is not the safest as we don't check for duplicates
			jsonVars = append(jsonVars, fmt.Sprintf("%q:%q", v.key, v.value))
		default:
			sb.WriteString(fmt.Sprintf("%s=%s\n", v.key, v.value))
		}
	}

	if format == FormatJSON {
		sb.WriteString("{")
		sb.WriteString(strings.Join(jsonVars, ","))
		sb.WriteString("}")
	}

	output := sb.String()
	if !opts.Write {
		fmt.Println(output)
		return nil
	}

	if err := os.WriteFile(file, []byte(output), 0o644); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
}

func run(ctx context.Context, opts runOpts, args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing executable")
	}

	exe := args[0]

	file := opts.File
	if opts.Name != "" {
		file = fmt.Sprintf("%s.%s", file, opts.Name)
	}

	// TODO: Move out
	key, err := loadKey()
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	vars, err := loadDecryptedEnv(file, key)
	if err != nil {
		return fmt.Errorf("error loading env file: %w", err)
	}

	for _, v := range vars {
		if err := os.Setenv(v.key, v.value); err != nil {
			return fmt.Errorf("error setting env var %s: %w", v.key, err)
		}
	}

	// Execute the new process in place of the Go process
	exe, err = exec.LookPath(exe)
	if err != nil {
		return fmt.Errorf("executable not found: %s", exe)
	}

	// TODO: Resolve shell alias
	/*
	  out, err := exec.Command("bash", "-i", "-c", "type "+exe).Output()
	  // Example output: alias gs='git status'
	  aliasPrefix := "alias " + exe + "="
	  if strings.HasPrefix(outputStr, aliasPrefix) {
	    aliasVal := strings.Trim(strings.TrimPrefix(outputStr, aliasPrefix), "'\" \n")
	    // aliasVal now contains the resolved command, e.g., "git status"
	  }
	*/

	err = syscall.Exec(exe, args, os.Environ())
	if err != nil {
		fmt.Println("Error executing process:", err, exe, args)
		os.Exit(1)
	}
	return nil
}
