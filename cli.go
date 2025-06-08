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

// TODO: Add support for subcommand and option autocomplete, as well as env keys
// Load from JSON into env :shrug:

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

type addOpts struct {
	Name    string
	File    string
	FmtOpts *fmtOpts
	print   bool
}

type setOpts struct {
	Name    string
	File    string
	FmtOpts *fmtOpts
	print   bool
}

type getOpts struct {
	Name       string
	File       string
	FmtOpts    *fmtOpts
	ValuesOnly bool
}

type getVOpts struct {
	Name      string
	File      string
	Separator string
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

	addCmd := new(command[addOpts])
	addCmd.flags = flag.NewFlagSet("add", flag.ExitOnError)
	addCmd.flags.StringVarP(&addCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	addCmd.flags.StringVarP(&addCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	addCmd.val.FmtOpts = NewFmtOpts(addCmd.flags)
	addCmd.flags.BoolVarP(&addCmd.val.print, "print", "p", false, "Prints the output instead of writing to the file.")
	addCmd.fn = addCmdFn
	cmds[addCmd.flags.Name()] = addCmd

	setCmd := new(command[setOpts])
	setCmd.flags = flag.NewFlagSet("set", flag.ExitOnError)
	setCmd.flags.StringVarP(&setCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	setCmd.flags.StringVarP(&setCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	setCmd.val.FmtOpts = NewFmtOpts(setCmd.flags)
	setCmd.flags.BoolVarP(&setCmd.val.print, "print", "p", false, "Prints the output instead of writing to the file.")
	setCmd.fn = setCmdFn
	cmds[setCmd.flags.Name()] = setCmd

	getCmd := new(command[getOpts])
	getCmd.flags = flag.NewFlagSet("get", flag.ExitOnError)
	getCmd.flags.StringVarP(&getCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	getCmd.flags.StringVarP(&getCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	getCmd.flags.BoolVarP(&getCmd.val.ValuesOnly, "vals", "v", false, "Prints only the values without keys. Use getv command instead to set a custom separator. Ignores formatting options.")
	getCmd.val.FmtOpts = NewFmtOpts(getCmd.flags)
	getCmd.fn = getCmdFn
	cmds[getCmd.flags.Name()] = getCmd

	getVCmd := new(command[getVOpts])
	getVCmd.flags = flag.NewFlagSet("getv", flag.ExitOnError)
	getVCmd.flags.StringVarP(&getVCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	getVCmd.flags.StringVarP(&getVCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	getVCmd.flags.StringVarP(&getVCmd.val.Separator, "separator", "s", "\n", "Separator for the values (default is new line)")
	getVCmd.fn = getVCmdFn
	cmds[getVCmd.flags.Name()] = getVCmd

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

func getVCmdFn(ctx context.Context, opts getVOpts, args ...string) error {
	file := opts.File
	if opts.Name != "" {
		file = fmt.Sprintf("%s.%s", file, opts.Name)
	}

	key, err := loadKey()
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	vars, err := loadDecryptedEnv(ctx, file, key)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	varKeys := make(map[string]envVar, len(vars))
	for _, v := range vars {
		varKeys[v.key] = v
	}

	vals := make([]string, 0, len(vars))
	if len(args) == 0 {
		for _, v := range vars {
			vals = append(vals, v.value)
		}
		fmt.Println(strings.Join(vals, opts.Separator))
		return nil
	}

	for _, arg := range args {
		if v, exists := varKeys[arg]; exists {
			vals = append(vals, v.value)
		} else {
			return fmt.Errorf("variable %s not found in %s file", arg, file)
		}
	}
	fmt.Println(strings.Join(vals, opts.Separator))
	return nil
}

func getCmdFn(ctx context.Context, opts getOpts, args ...string) error {
	if opts.ValuesOnly {
		return getVCmdFn(ctx, getVOpts{opts.Name, opts.File, "\n"}, args...)
	}

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

	vars, err := loadDecryptedEnv(ctx, file, key)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	varKeys := make(map[string]envVar, len(vars))
	for _, v := range vars {
		varKeys[v.key] = v
	}

	if len(args) == 0 {
		for _, v := range vars {
			switch format {
			case FormatJSON:
				fmt.Printf("%q:%q\n", v.key, v.value)
			default:
				fmt.Printf("%s=%s\n", v.key, v.value)
			}
		}
		return nil
	}

	for _, arg := range args {
		if v, exists := varKeys[arg]; exists {
			switch format {
			case FormatJSON:
				fmt.Printf("%q:%q\n", v.key, v.value)
			default:
				fmt.Printf("%s=%s\n", v.key, v.value)
			}
		} else {
			return fmt.Errorf("variable %s not found in %s file", arg, file)
		}
	}
	return nil
}

func setCmdFn(ctx context.Context, opts setOpts, args ...string) error {
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

	vars, err := loadEnv(ctx, file)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	varKeys := make(map[string]int, len(vars))
	for i, v := range vars {
		varKeys[v.key] = i
	}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			ciphertext, err := encrypt(v, key)
			if err != nil {
				return fmt.Errorf("error encrypting value: %w", err)
			}
			if idx, exists := varKeys[k]; exists {
				vars[idx].value = ciphertext
			} else {
				vars = append(vars, envVar{key: k, value: ciphertext})
			}
		} else {
			return fmt.Errorf("invalid argument format: %s, expected key=value", arg)
		}
	}

	var sb strings.Builder
	jsonVars := make([]string, 0, len(vars))
	for _, v := range vars {
		switch format {
		case FormatJSON:
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
	if opts.print {
		fmt.Println(output)
		return nil
	}

	if err := os.WriteFile(file, []byte(output), 0o644); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
}

func addCmdFn(ctx context.Context, opts addOpts, args ...string) error {
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

	vars, err := loadEnv(ctx, file)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	varKeys := make(map[string]bool, len(vars))
	for _, v := range vars {
		varKeys[v.key] = true
	}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			k := strings.TrimSpace(parts[0])
			if varKeys[k] {
				return fmt.Errorf("variable %s already exists in %s file", k, file)
			}

			v := strings.TrimSpace(parts[1])
			ciphertext, err := encrypt(v, key)
			if err != nil {
				return fmt.Errorf("error encrypting value: %w", err)
			}
			vars = append(vars, envVar{key: k, value: ciphertext})
		} else {
			return fmt.Errorf("invalid argument format: %s, expected key=value", arg)
		}
	}

	var sb strings.Builder
	jsonVars := make([]string, 0, len(vars))
	for _, v := range vars {
		switch format {
		case FormatJSON:
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
	if opts.print {
		fmt.Println(output)
		return nil
	}

	if err := os.WriteFile(file, []byte(output), 0o644); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
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
	vars, err := loadEnv(ctx, file)
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
	vars, err := loadEnv(ctx, file)
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

	vars, err := loadDecryptedEnv(ctx, file, key)
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
