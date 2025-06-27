package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/almahoozi/envx/pkg/config"
	"github.com/almahoozi/envx/pkg/crypto"
	"github.com/almahoozi/envx/pkg/env"
	flag "github.com/spf13/pflag"
	"golang.org/x/term"
)

// TODO: Add support for subcommand and option autocomplete, as well as env keys
// Load from JSON into env :shrug:

var emptyPassword = string([]byte{1})

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

// Use the Format type from the env package
type Format = env.Format

const (
	FormatEnv  = env.FormatEnv
	FormatJSON = env.FormatJSON
	FormatYAML = env.FormatYAML
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
	Name     string
	File     string
	KeyStore string
	Password string
	FmtOpts  *fmtOpts
	Write    bool
}

type decryptOpts struct {
	Name     string
	File     string
	KeyStore string
	Password string
	FmtOpts  *fmtOpts
	Write    bool
}

type addOpts struct {
	Name     string
	File     string
	KeyStore string
	Password string
	FmtOpts  *fmtOpts
	print    bool
}

type setOpts struct {
	Name     string
	File     string
	KeyStore string
	Password string
	FmtOpts  *fmtOpts
	print    bool
}

type getOpts struct {
	Name       string
	File       string
	KeyStore   string
	Password   string
	FmtOpts    *fmtOpts
	ValuesOnly bool
}

type getVOpts struct {
	Name      string
	File      string
	KeyStore  string
	Password  string
	Separator string
}

type runOpts struct {
	Name     string
	File     string
	KeyStore string
	Password string
	Args     []string
}

type configOpts struct {
	Directory bool
	Separator string
}

type executor interface {
	execute(ctx context.Context, args ...string) error
}

// applyConfigToRunOpts merges configuration with CLI options for run command
func applyConfigToRunOpts(opts *runOpts) error {
	manager := config.NewManager()
	cfg, err := manager.GetEffectiveConfig()
	if err != nil {
		return err
	}

	// Apply config defaults only if CLI flags weren't explicitly set
	if opts.File == "" {
		opts.File = cfg.File
	}
	if opts.Name == "" {
		opts.Name = cfg.Name
	}
	if opts.KeyStore == "" {
		opts.KeyStore = cfg.Keystore
	}

	return nil
}

// applyConfigToGetOpts merges configuration with CLI options for get command
func applyConfigToGetOpts(opts *getOpts) error {
	manager := config.NewManager()
	cfg, err := manager.GetEffectiveConfig()
	if err != nil {
		return err
	}

	// Apply config defaults only if CLI flags weren't explicitly set
	if opts.File == "" {
		opts.File = cfg.File
	}
	if opts.Name == "" {
		opts.Name = cfg.Name
	}
	if opts.KeyStore == "" {
		opts.KeyStore = cfg.Keystore
	}

	return nil
}

// applyConfigToGetVOpts merges configuration with CLI options for getv command
func applyConfigToGetVOpts(opts *getVOpts) error {
	manager := config.NewManager()
	cfg, err := manager.GetEffectiveConfig()
	if err != nil {
		return err
	}

	// Apply config defaults only if CLI flags weren't explicitly set
	if opts.File == "" {
		opts.File = cfg.File
	}
	if opts.Name == "" {
		opts.Name = cfg.Name
	}
	if opts.KeyStore == "" {
		opts.KeyStore = cfg.Keystore
	}

	return nil
}

// applyConfigToSetOpts merges configuration with CLI options for set command
func applyConfigToSetOpts(opts *setOpts) error {
	manager := config.NewManager()
	cfg, err := manager.GetEffectiveConfig()
	if err != nil {
		return err
	}

	// Apply config defaults only if CLI flags weren't explicitly set
	if opts.File == "" {
		opts.File = cfg.File
	}
	if opts.Name == "" {
		opts.Name = cfg.Name
	}
	if opts.KeyStore == "" {
		opts.KeyStore = cfg.Keystore
	}

	return nil
}

// applyConfigToAddOpts merges configuration with CLI options for add command
func applyConfigToAddOpts(opts *addOpts) error {
	manager := config.NewManager()
	cfg, err := manager.GetEffectiveConfig()
	if err != nil {
		return err
	}

	// Apply config defaults only if CLI flags weren't explicitly set
	if opts.File == "" {
		opts.File = cfg.File
	}
	if opts.Name == "" {
		opts.Name = cfg.Name
	}
	if opts.KeyStore == "" {
		opts.KeyStore = cfg.Keystore
	}

	return nil
}

// applyConfigToEncryptOpts merges configuration with CLI options for encrypt command
func applyConfigToEncryptOpts(opts *encryptOpts) error {
	manager := config.NewManager()
	cfg, err := manager.GetEffectiveConfig()
	if err != nil {
		return err
	}

	// Apply config defaults only if CLI flags weren't explicitly set
	if opts.File == "" {
		opts.File = cfg.File
	}
	if opts.Name == "" {
		opts.Name = cfg.Name
	}
	if opts.KeyStore == "" {
		opts.KeyStore = cfg.Keystore
	}

	return nil
}

// applyConfigToDecryptOpts merges configuration with CLI options for decrypt command
func applyConfigToDecryptOpts(opts *decryptOpts) error {
	manager := config.NewManager()
	cfg, err := manager.GetEffectiveConfig()
	if err != nil {
		return err
	}

	// Apply config defaults only if CLI flags weren't explicitly set
	if opts.File == "" {
		opts.File = cfg.File
	}
	if opts.Name == "" {
		opts.Name = cfg.Name
	}
	if opts.KeyStore == "" {
		opts.KeyStore = cfg.Keystore
	}

	return nil
}

// resolveFileWithConfig uses the config system's smart file resolution
func resolveFileWithConfig(explicitFile, explicitName string) (string, error) {
	manager := config.NewManager()
	return manager.ResolveFile(explicitFile, explicitName)
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
	runCmd.flags.StringVarP(&runCmd.val.File, "file", "f", "", "Uses a specific file instead of the default .env")
	runCmd.flags.StringVarP(&runCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	runCmd.flags.StringVarP(&runCmd.val.KeyStore, "keystore", "k", "", "Keystore type to use (macos, password, mock)")
	runCmd.flags.StringVarP(&runCmd.val.Password, "password", "P", "", "Password for password-based keystore (implies --keystore password; use ENVX_PASSWORD env var for better security)")
	runCmd.flags.Lookup("password").NoOptDefVal = emptyPassword
	runCmd.fn = run
	cmds[runCmd.flags.Name()] = runCmd

	encCmd := new(command[encryptOpts])
	encCmd.flags = flag.NewFlagSet("encrypt", flag.ExitOnError)
	encCmd.flags.StringVarP(&encCmd.val.File, "file", "f", "", "Uses a specific file instead of the default .env")
	encCmd.flags.StringVarP(&encCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	encCmd.flags.StringVarP(&encCmd.val.KeyStore, "keystore", "k", "", "Keystore type to use (macos, password, mock)")
	encCmd.flags.StringVarP(&encCmd.val.Password, "password", "P", "", "Password for password-based keystore (implies --keystore password; use ENVX_PASSWORD env var for better security)")
	encCmd.flags.Lookup("password").NoOptDefVal = emptyPassword
	encCmd.val.FmtOpts = NewFmtOpts(encCmd.flags)
	encCmd.flags.BoolVarP(&encCmd.val.Write, "write", "w", false, "Overwrites the file with encrypted values.")
	encCmd.fn = encryptCmd
	cmds[encCmd.flags.Name()] = encCmd

	decCmd := new(command[decryptOpts])
	decCmd.flags = flag.NewFlagSet("decrypt", flag.ExitOnError)
	decCmd.flags.StringVarP(&decCmd.val.File, "file", "f", "", "Uses a specific file instead of the default .env")
	decCmd.flags.StringVarP(&decCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	decCmd.flags.StringVarP(&decCmd.val.KeyStore, "keystore", "k", "", "Keystore type to use (macos, password, mock)")
	decCmd.flags.StringVarP(&decCmd.val.Password, "password", "P", "", "Password for password-based keystore (implies --keystore password; use ENVX_PASSWORD env var for better security)")
	decCmd.flags.Lookup("password").NoOptDefVal = emptyPassword
	decCmd.val.FmtOpts = NewFmtOpts(decCmd.flags)
	decCmd.flags.BoolVarP(&decCmd.val.Write, "write", "w", false, "Overwrites the file with decrypted values.")
	decCmd.fn = decryptCmd
	cmds[decCmd.flags.Name()] = decCmd

	addCmd := new(command[addOpts])
	addCmd.flags = flag.NewFlagSet("add", flag.ExitOnError)
	addCmd.flags.StringVarP(&addCmd.val.File, "file", "f", "", "Uses a specific file instead of the default .env")
	addCmd.flags.StringVarP(&addCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	addCmd.flags.StringVarP(&addCmd.val.KeyStore, "keystore", "k", "", "Keystore type to use (macos, password, mock)")
	addCmd.flags.StringVarP(&addCmd.val.Password, "password", "P", "", "Password for password-based keystore (implies --keystore password; use ENVX_PASSWORD env var for better security)")
	addCmd.flags.Lookup("password").NoOptDefVal = emptyPassword
	addCmd.val.FmtOpts = NewFmtOpts(addCmd.flags)
	addCmd.flags.BoolVarP(&addCmd.val.print, "print", "p", false, "Prints the output instead of writing to the file.")
	addCmd.fn = addCmdFn
	cmds[addCmd.flags.Name()] = addCmd

	setCmd := new(command[setOpts])
	setCmd.flags = flag.NewFlagSet("set", flag.ExitOnError)
	setCmd.flags.StringVarP(&setCmd.val.File, "file", "f", "", "Uses a specific file instead of the default .env")
	setCmd.flags.StringVarP(&setCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	setCmd.flags.StringVarP(&setCmd.val.KeyStore, "keystore", "k", "", "Keystore type to use (macos, password, mock)")
	setCmd.flags.StringVarP(&setCmd.val.Password, "password", "P", "", "Password for password-based keystore (implies --keystore password; use ENVX_PASSWORD env var for better security)")
	setCmd.flags.Lookup("password").NoOptDefVal = emptyPassword
	setCmd.val.FmtOpts = NewFmtOpts(setCmd.flags)
	setCmd.flags.BoolVarP(&setCmd.val.print, "print", "p", false, "Prints the output instead of writing to the file.")
	setCmd.fn = setCmdFn
	cmds[setCmd.flags.Name()] = setCmd

	getCmd := new(command[getOpts])
	getCmd.flags = flag.NewFlagSet("get", flag.ExitOnError)
	getCmd.flags.StringVarP(&getCmd.val.File, "file", "f", "", "Uses a specific file instead of the default .env")
	getCmd.flags.StringVarP(&getCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	getCmd.flags.StringVarP(&getCmd.val.KeyStore, "keystore", "k", "", "Keystore type to use (macos, password, mock)")
	getCmd.flags.StringVarP(&getCmd.val.Password, "password", "P", "", "Password for password-based keystore (implies --keystore password; use ENVX_PASSWORD env var for better security)")
	getCmd.flags.Lookup("password").NoOptDefVal = emptyPassword
	getCmd.flags.BoolVarP(&getCmd.val.ValuesOnly, "vals", "v", false, "Prints only the values without keys. Use getv command instead to set a custom separator. Ignores formatting options.")
	getCmd.val.FmtOpts = NewFmtOpts(getCmd.flags)
	getCmd.fn = getCmdFn
	cmds[getCmd.flags.Name()] = getCmd

	getVCmd := new(command[getVOpts])
	getVCmd.flags = flag.NewFlagSet("getv", flag.ExitOnError)
	getVCmd.flags.StringVarP(&getVCmd.val.File, "file", "f", "", "Uses a specific file instead of the default .env")
	getVCmd.flags.StringVarP(&getVCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	getVCmd.flags.StringVarP(&getVCmd.val.KeyStore, "keystore", "k", "", "Keystore type to use (macos, password, mock)")
	getVCmd.flags.StringVarP(&getVCmd.val.Password, "password", "P", "", "Password for password-based keystore (implies --keystore password; use ENVX_PASSWORD env var for better security)")
	getVCmd.flags.Lookup("password").NoOptDefVal = emptyPassword
	getVCmd.flags.StringVarP(&getVCmd.val.Separator, "separator", "s", "\n", "Separator for the values (default is new line)")
	getVCmd.fn = getVCmdFn
	cmds[getVCmd.flags.Name()] = getVCmd

	// Config commands
	configCmd := new(command[configOpts])
	configCmd.flags = flag.NewFlagSet("config", flag.ExitOnError)
	configCmd.flags.BoolVarP(&configCmd.val.Directory, "directory", "d", false, "Operate on directory-level configuration (.envx.yaml)")
	configCmd.flags.StringVarP(&configCmd.val.Separator, "separator", "s", "\n", "Separator for the values (used with getv subcommand, default is new line)")
	configCmd.fn = configCmdFn
	cmds[configCmd.flags.Name()] = configCmd

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
	// Store original CLI values before applying config
	originalFile := opts.File
	originalName := opts.Name

	// Apply configuration defaults
	if err := applyConfigToGetVOpts(&opts); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Use smart file resolution with original CLI values
	file, err := resolveFileWithConfig(originalFile, originalName)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}

	key, err := loadKeyWithStringTypeAndPassword(opts.KeyStore, opts.Password)
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	encryptor := crypto.NewAESEncryptor()
	vars, err := loadDecryptedEnv(ctx, file, encryptor, key)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	varMap := vars.ToMap()

	vals := make([]string, 0, len(vars))
	if len(args) == 0 {
		for _, v := range vars {
			vals = append(vals, v.Value)
		}
		fmt.Println(strings.Join(vals, opts.Separator))
		return nil
	}

	for _, arg := range args {
		if value, exists := varMap[arg]; exists {
			vals = append(vals, value)
		} else {
			return fmt.Errorf("variable %s not found in %s file", arg, file)
		}
	}
	fmt.Println(strings.Join(vals, opts.Separator))
	return nil
}

func getCmdFn(ctx context.Context, opts getOpts, args ...string) error {
	// Store original CLI values before applying config
	originalFile := opts.File
	originalName := opts.Name

	// Apply configuration defaults
	if err := applyConfigToGetOpts(&opts); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	if opts.ValuesOnly {
		return getVCmdFn(ctx, getVOpts{opts.Name, opts.File, opts.KeyStore, opts.Password, "\n"}, args...)
	}

	format, err := opts.FmtOpts.Format()
	if err != nil {
		return fmt.Errorf("error parsing format: %w", err)
	}
	if format == FormatYAML {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Use smart file resolution with original CLI values
	file, err := resolveFileWithConfig(originalFile, originalName)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}

	key, err := loadKeyWithStringTypeAndPassword(opts.KeyStore, opts.Password)
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	encryptor := crypto.NewAESEncryptor()
	vars, err := loadDecryptedEnv(ctx, file, encryptor, key)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	varMap := vars.ToMap()

	if len(args) == 0 {
		for _, v := range vars {
			switch format {
			case FormatJSON:
				fmt.Printf("%q:%q\n", v.Key, v.Value)
			default:
				fmt.Printf("%s=%s\n", v.Key, v.Value)
			}
		}
		return nil
	}

	for _, arg := range args {
		if value, exists := varMap[arg]; exists {
			switch format {
			case FormatJSON:
				fmt.Printf("%q:%q\n", arg, value)
			default:
				fmt.Printf("%s=%s\n", arg, value)
			}
		} else {
			return fmt.Errorf("variable %s not found in %s file", arg, file)
		}
	}
	return nil
}

func setCmdFn(ctx context.Context, opts setOpts, args ...string) error {
	// Store original CLI values before applying config
	originalFile := opts.File
	originalName := opts.Name

	// Apply configuration defaults
	if err := applyConfigToSetOpts(&opts); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	format, err := opts.FmtOpts.Format()
	if err != nil {
		return fmt.Errorf("error parsing format: %w", err)
	}
	if format == FormatYAML {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Use smart file resolution with original CLI values
	file, err := resolveFileWithConfig(originalFile, originalName)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}

	key, err := loadKeyWithStringTypeAndPassword(opts.KeyStore, opts.Password)
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	vars, err := loadEnv(ctx, file)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	// Parse arguments supporting both key=value and key-only formats
	keyValues, err := parseKeyValueArgs(args)
	if err != nil {
		return fmt.Errorf("error parsing arguments: %w", err)
	}

	encryptor := crypto.NewAESEncryptor()

	if opts.print {
		// Create a new Variables slice with only the newly set values
		newVars := make(env.Variables, 0, len(keyValues))
		for k, v := range keyValues {
			ciphertext, err := encryptor.Encrypt(v, key)
			if err != nil {
				return fmt.Errorf("error encrypting value for key %s: %w", k, err)
			}
			newVars = append(newVars, env.Variable{Key: k, Value: ciphertext})
		}

		writer := env.NewFileWriter()
		// Use a temp file to get the output
		tempFile, err := createTempFile()
		if err != nil {
			return err
		}
		err = writer.Write(tempFile, newVars, format)
		if err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("error reading formatted output: %w", err)
		}
		removeFileIgnoreError(tempFile)
		fmt.Print(string(content))
		return nil
	}

	// If not printing, update the actual vars and write to file
	for k, v := range keyValues {
		ciphertext, err := encryptor.Encrypt(v, key)
		if err != nil {
			return fmt.Errorf("error encrypting value for key %s: %w", k, err)
		}
		vars.Set(k, ciphertext)
	}

	writer := env.NewFileWriter()
	if err := writer.Write(file, vars, format); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
}

func addCmdFn(ctx context.Context, opts addOpts, args ...string) error {
	// Store original CLI values before applying config
	originalFile := opts.File
	originalName := opts.Name

	// Apply configuration defaults
	if err := applyConfigToAddOpts(&opts); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	format, err := opts.FmtOpts.Format()
	if err != nil {
		return fmt.Errorf("error parsing format: %w", err)
	}
	if format == FormatYAML {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Use smart file resolution with original CLI values
	file, err := resolveFileWithConfig(originalFile, originalName)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}

	key, err := loadKeyWithStringTypeAndPassword(opts.KeyStore, opts.Password)
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	vars, err := loadEnv(ctx, file)
	if err != nil {
		return fmt.Errorf("error loading %s file: %w", file, err)
	}

	varMap := vars.ToMap()

	// Parse arguments supporting both key=value and key-only formats
	keyValues, err := parseKeyValueArgs(args)
	if err != nil {
		return fmt.Errorf("error parsing arguments: %w", err)
	}

	// Check for existing keys first
	for k := range keyValues {
		if _, exists := varMap[k]; exists {
			return fmt.Errorf("variable %s already exists in %s file", k, file)
		}
	}

	encryptor := crypto.NewAESEncryptor()

	if opts.print {
		// Create a new Variables slice with only the newly added values
		newVars := make(env.Variables, 0, len(keyValues))
		for k, v := range keyValues {
			ciphertext, err := encryptor.Encrypt(v, key)
			if err != nil {
				return fmt.Errorf("error encrypting value for key %s: %w", k, err)
			}
			newVars = append(newVars, env.Variable{Key: k, Value: ciphertext})
		}

		writer := env.NewFileWriter()
		// Use a temp file to get the output
		tempFile, err := createTempFile()
		if err != nil {
			return err
		}
		err = writer.Write(tempFile, newVars, format)
		if err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("error reading formatted output: %w", err)
		}
		removeFileIgnoreError(tempFile)
		fmt.Print(string(content))
		return nil
	}

	// If not printing, update the actual vars and write to file
	for k, v := range keyValues {
		ciphertext, err := encryptor.Encrypt(v, key)
		if err != nil {
			return fmt.Errorf("error encrypting value for key %s: %w", k, err)
		}
		vars.Set(k, ciphertext)
	}

	writer := env.NewFileWriter()
	if err := writer.Write(file, vars, format); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
}

func encryptCmd(ctx context.Context, opts encryptOpts, args ...string) error {
	// Store original CLI values before applying config
	originalFile := opts.File
	originalName := opts.Name

	// Apply configuration defaults
	if err := applyConfigToEncryptOpts(&opts); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	format, err := opts.FmtOpts.Format()
	if err != nil {
		return fmt.Errorf("error parsing format: %w", err)
	}
	if format == FormatYAML {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Use smart file resolution with original CLI values
	file, err := resolveFileWithConfig(originalFile, originalName)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}

	key, err := loadKeyWithStringTypeAndPassword(opts.KeyStore, opts.Password)
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

	encryptor := crypto.NewAESEncryptor()

	for i, v := range vars {
		// If it isn't already encrypted, encrypt it
		if len(args) == 0 || argMap[v.Key] {
			ciphertext, err := encryptor.Encrypt(v.Value, key)
			if err != nil {
				return fmt.Errorf("error encrypting value: %w", err)
			}
			vars[i].Value = ciphertext
		}
	}

	if !opts.Write {
		writer := env.NewFileWriter()
		// Use a temp file to get the output
		tempFile, err := createTempFile()
		if err != nil {
			return err
		}
		err = writer.Write(tempFile, vars, format)
		if err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("error reading formatted output: %w", err)
		}
		removeFileIgnoreError(tempFile)
		fmt.Print(string(content))
		return nil
	}

	writer := env.NewFileWriter()
	if err := writer.Write(file, vars, format); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
}

func decryptCmd(ctx context.Context, opts decryptOpts, args ...string) error {
	// Store original CLI values before applying config
	originalFile := opts.File
	originalName := opts.Name

	// Apply configuration defaults
	if err := applyConfigToDecryptOpts(&opts); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	format, err := opts.FmtOpts.Format()
	if err != nil {
		return fmt.Errorf("error parsing format: %w", err)
	}
	if format == FormatYAML {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Use smart file resolution with original CLI values
	file, err := resolveFileWithConfig(originalFile, originalName)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}

	key, err := loadKeyWithStringTypeAndPassword(opts.KeyStore, opts.Password)
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

	encryptor := crypto.NewAESEncryptor()

	for i, v := range vars {
		if len(args) == 0 || argMap[v.Key] {
			plaintext, err := encryptor.Decrypt(v.Value, key)
			if err != nil {
				return fmt.Errorf("error decrypting value: %w", err)
			}
			vars[i].Value = plaintext
		}
	}

	if !opts.Write {
		writer := env.NewFileWriter()
		// Use a temp file to get the output
		tempFile, err := createTempFile()
		if err != nil {
			return err
		}
		err = writer.Write(tempFile, vars, format)
		if err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("error reading formatted output: %w", err)
		}
		removeFileIgnoreError(tempFile)
		fmt.Print(string(content))
		return nil
	}

	writer := env.NewFileWriter()
	if err := writer.Write(file, vars, format); err != nil {
		return fmt.Errorf("error writing %s file: %w", file, err)
	}
	return nil
}

func run(ctx context.Context, opts runOpts, args ...string) error {
	// Store original CLI values before applying config
	originalFile := opts.File
	originalName := opts.Name

	// Apply configuration defaults
	if err := applyConfigToRunOpts(&opts); err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	if len(args) < 1 {
		return fmt.Errorf("missing executable")
	}

	exe := args[0]

	// Use smart file resolution with original CLI values
	file, err := resolveFileWithConfig(originalFile, originalName)
	if err != nil {
		return fmt.Errorf("error resolving file: %w", err)
	}

	// TODO: Move out
	key, err := loadKeyWithStringTypeAndPassword(opts.KeyStore, opts.Password)
	if err != nil {
		return fmt.Errorf("error loading key: %w", err)
	}

	encryptor := crypto.NewAESEncryptor()
	vars, err := loadDecryptedEnv(ctx, file, encryptor, key)
	if err != nil {
		return fmt.Errorf("error loading env file: %w", err)
	}

	for _, v := range vars {
		if err := os.Setenv(v.Key, v.Value); err != nil {
			return fmt.Errorf("error setting env var %s: %w", v.Key, err)
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

	err = syscall.Exec(exe, args, os.Environ()) // #nosec G204 -- Intentional subprocess execution with validated executable path
	if err != nil {
		fmt.Println("Error executing process:", err, exe, args)
		os.Exit(1)
	}
	return nil
}

// removeFileIgnoreError removes a file and ignores any error (for temp file cleanup)
func removeFileIgnoreError(filename string) {
	_ = os.Remove(filename) // Ignore error for temp file cleanup
}

// createTempFile creates a temporary file and returns its name
func createTempFile() (string, error) {
	tempFile, err := os.CreateTemp("", "envx_temp_output_*")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	return tempFileName, nil
}

// promptForSecretValue prompts the user to enter a secret value securely
func promptForSecretValue(key string) (string, error) {
	fmt.Printf("Enter value for %s: ", key)

	// Read password without echoing to terminal
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Println() // Print newline after password input
	return string(bytePassword), nil
}

// parseKeyValueArgs parses arguments that can be either "key=value" or just "key"
// For keys without values, it prompts securely for the value
func parseKeyValueArgs(args []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// Handle key=value format
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" {
				return nil, fmt.Errorf("empty key in argument: %s", arg)
			}
			result[key] = value
		} else {
			// Handle key-only format - prompt for value
			key := strings.TrimSpace(arg)
			if key == "" {
				return nil, fmt.Errorf("empty key: %s", arg)
			}

			value, err := promptForSecretValue(key)
			if err != nil {
				return nil, fmt.Errorf("failed to get value for key %s: %w", key, err)
			}
			result[key] = value
		}
	}

	return result, nil
}

func configCmdFn(ctx context.Context, opts configOpts, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("config command requires a subcommand (set, get, getv, list, show, init, reset)")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "set":
		return configSetCmd(ctx, opts, subArgs...)
	case "get":
		return configGetCmd(ctx, opts, subArgs...)
	case "getv":
		return configGetVCmd(ctx, opts, subArgs...)
	case "list":
		return configListCmd(ctx, opts, subArgs...)
	case "show":
		return configShowCmd(ctx, opts, subArgs...)
	case "init":
		return configInitCmd(ctx, opts, subArgs...)
	case "reset":
		return configResetCmd(ctx, opts, subArgs...)
	default:
		return fmt.Errorf("unknown config subcommand: %s (supported: set, get, getv, list, show, init, reset)", subcommand)
	}
}

func configSetCmd(ctx context.Context, opts configOpts, args ...string) error {
	if len(args) != 2 {
		return fmt.Errorf("config set requires exactly 2 arguments: key value")
	}

	key, value := args[0], args[1]
	manager := config.NewManager()

	if err := manager.Set(key, value, opts.Directory); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	target := "global"
	if opts.Directory {
		target = "directory"
	}
	fmt.Printf("Set %s config: %s = %s\n", target, key, value)
	return nil
}

func configGetCmd(ctx context.Context, opts configOpts, args ...string) error {
	manager := config.NewManager()

	if len(args) == 0 {
		// Get all configuration values
		report, err := manager.GetReport()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		fmt.Printf("keystore = %s (from %s)\n", report.Keystore.Value, report.Keystore.Source)
		fmt.Printf("file = %s (from %s)\n", report.File.Value, report.File.Source)
		fmt.Printf("name = %s (from %s)\n", report.Name.Value, report.Name.Source)
		fmt.Printf("format = %s (from %s)\n", report.Format.Value, report.Format.Source)
		fmt.Printf("key_name = %s (from %s)\n", report.KeyName.Value, report.KeyName.Source)
		fmt.Printf("file_resolution = %s (from %s)\n", strings.Join(report.FileResolution.Value, ","), report.FileResolution.Source)
		fmt.Printf("backup_on_write = %s (from %s)\n", report.BackupOnWrite.Value, report.BackupOnWrite.Source)
		return nil
	}

	// Get specific key
	key := args[0]

	// Check if it's an array field
	if strings.ToLower(key) == "file_resolution" || strings.ToLower(key) == "fileresolution" {
		values, source, err := manager.GetArray(key)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}
		fmt.Printf("%s = %s (from %s)\n", key, strings.Join(values, ","), source)
		return nil
	}

	// Regular string field
	value, source, err := manager.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	fmt.Printf("%s = %s (from %s)\n", key, value, source)
	return nil
}

func configListCmd(ctx context.Context, opts configOpts, args ...string) error {
	manager := config.NewManager()

	globalPath, dirPath, err := manager.GetConfigPaths()
	if err != nil {
		return fmt.Errorf("failed to get config paths: %w", err)
	}

	globalExists, dirExists, err := manager.ConfigExists()
	if err != nil {
		return fmt.Errorf("failed to check config existence: %w", err)
	}

	fmt.Printf("Configuration files:\n")
	fmt.Printf("  Global: %s", globalPath)
	if globalExists {
		fmt.Printf(" (exists)\n")
	} else {
		fmt.Printf(" (not found)\n")
	}

	fmt.Printf("  Directory: %s", dirPath)
	if dirExists {
		fmt.Printf(" (exists)\n")
	} else {
		fmt.Printf(" (not found)\n")
	}

	fmt.Printf("\nEffective configuration:\n")
	return configGetCmd(ctx, opts, args...)
}

func configShowCmd(ctx context.Context, opts configOpts, args ...string) error {
	manager := config.NewManager()

	report, err := manager.GetReport()
	if err != nil {
		return fmt.Errorf("failed to get config report: %w", err)
	}

	// Output as JSON for structured viewing
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config report: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func configInitCmd(ctx context.Context, opts configOpts, args ...string) error {
	manager := config.NewManager()

	if err := manager.Init(opts.Directory); err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}

	target := "global"
	if opts.Directory {
		target = "directory"
	}
	fmt.Printf("Initialized %s configuration (reset to defaults)\n", target)
	return nil
}

func configResetCmd(ctx context.Context, opts configOpts, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("config reset requires exactly 1 argument: key")
	}

	key := args[0]
	manager := config.NewManager()

	if err := manager.Reset(key, opts.Directory); err != nil {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	target := "global"
	if opts.Directory {
		target = "directory"
	}
	fmt.Printf("Reset %s config key: %s\n", target, key)
	return nil
}

func configGetVCmd(ctx context.Context, opts configOpts, args ...string) error {
	manager := config.NewManager()

	if len(args) == 0 {
		// Get all configuration values
		report, err := manager.GetReport()
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		values := []string{
			report.Keystore.Value,
			report.File.Value,
			report.Name.Value,
			report.Format.Value,
			report.KeyName.Value,
			strings.Join(report.FileResolution.Value, ","),
			report.BackupOnWrite.Value,
		}
		fmt.Println(strings.Join(values, opts.Separator))
		return nil
	}

	// Get specific key values
	values := make([]string, 0, len(args))
	for _, key := range args {
		// Check if it's an array field
		if strings.ToLower(key) == "file_resolution" || strings.ToLower(key) == "fileresolution" {
			arrayValues, _, err := manager.GetArray(key)
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}
			values = append(values, strings.Join(arrayValues, ","))
		} else {
			// Regular string field
			value, _, err := manager.Get(key)
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}
			values = append(values, value)
		}
	}

	fmt.Println(strings.Join(values, opts.Separator))
	return nil
}
