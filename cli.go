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

type encryptOpts struct {
	Name  string
	File  string
	Write bool
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
	runCmd.flags.StringVarP(&runCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	runCmd.flags.StringVarP(&runCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	runCmd.fn = run
	cmds[runCmd.flags.Name()] = runCmd

	encCmd := new(command[encryptOpts])
	encCmd.flags = flag.NewFlagSet("encrypt", flag.ExitOnError)
	encCmd.flags.StringVarP(&encCmd.val.Name, "name", "n", "", "Looks for .env.<name> file instead of .env")
	encCmd.flags.StringVarP(&encCmd.val.File, "file", "f", ".env", "Uses a specific file instead of the default .env")
	encCmd.flags.BoolVarP(&encCmd.val.Write, "write", "w", false, "Overwrites the file with encrypted values.")
	encCmd.fn = encryptCmd
	cmds[encCmd.flags.Name()] = encCmd

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

	for i, v := range vars {
		// If it isn't already encrypted, encrypt it
		// Well, we decrypted everything, so just encrypt everything
		vars[i].value = encrypt(v.value, key)
	}

	var sb strings.Builder
	for _, v := range vars {
		sb.WriteString(fmt.Sprintf("%s=%s\n", v.key, v.value))
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

	vars, err := loadEnv(file, key)
	if err != nil {
		return fmt.Errorf("error loading env file: %w", err)
	}

	for _, v := range vars {
		os.Setenv(v.key, v.value)
	}

	// Execute the new process in place of the Go process
	exe, err = exec.LookPath(exe)
	if err != nil {
		return fmt.Errorf("executable not found: %s", exe)
	}

	err = syscall.Exec(exe, args, os.Environ())
	if err != nil {
		fmt.Println("Error executing process:", err, exe, args)
		os.Exit(1)
	}
	return nil
}
