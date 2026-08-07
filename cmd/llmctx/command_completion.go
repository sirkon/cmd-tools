package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"

	"github.com/sirkon/cmd-tools/cmd/llmctx/internal/complscript"
)

// CommandCompletion сгенерировать скрипт автодополнения.
type CommandCompletion struct {
	Shell ShellType `arg:""    help:"Shell type to generate completion script for."`
	Save  bool      `short:"s" help:"Save completion script for llmctx within completion files."`
}

func (c *CommandCompletion) Run(ctx *runContext) error {
	var complete string
	switch c.Shell {
	case shellTypeBash:
		complete = complscript.Bash
	case shellTypeZsh:
		complete = complscript.Zsh
	case shellTypeFish:
		complete = complscript.Fish
	}

	if !c.Save {
		fmt.Println(complete)
		return nil
	}

	complFileName := shellComplFileName(c.Shell, appName)
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get current user home directory")
	}

	var absComplRoot string
	switch c.Shell {
	case shellTypeBash:
		absComplRoot = filepath.Join(home, ".local", "share", "bash-completion")
	case shellTypeZsh:
		absComplRoot = filepath.Join(home, ".zsh")
	case shellTypeFish:
		cfgDir, err := os.UserConfigDir()
		if err != nil {
			return errors.Wrap(err, "get current user config directory")
		}
		absComplRoot = filepath.Join(cfgDir, "fish")
	}
	absComplFileName := filepath.Join(absComplRoot, "completions", complFileName)

	if err := os.MkdirAll(filepath.Dir(absComplFileName), 0755); err != nil {
		return errors.Wrap(err, "create completion directory")
	}
	if err := os.WriteFile(absComplFileName, []byte(complete), 0644); err != nil {
		return errors.Wrap(err, "write completion script")
	}

	message.Infof("Completion script saved to %s\n", absComplFileName)

	return nil
}

type ShellType int

const (
	shellTypeBash ShellType = iota + 1
	shellTypeZsh
	shellTypeFish
)

func (s *ShellType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "bash":
		*s = shellTypeBash
	case "zsh":
		*s = shellTypeZsh
	case "fish":
		*s = shellTypeFish
	default:
		return fmt.Errorf("unsupported shell type %q", string(text))
	}

	return nil
}

// shellComplFileName базовое имя файла.
func shellComplFileName(shell ShellType, name string) string {
	switch shell {
	case shellTypeZsh:
		return "_" + name
	case shellTypeBash:
		return name
	case shellTypeFish:
		return name + ".fish"
	default:
		return "unknown"
	}
}
