package main

import (
	"fmt"

	"github.com/sirkon/cmd-tools/cmd/llmctx/internal/complscript"
)

// CommandCompletion сгенерировать скрипт автодополнения.
type CommandCompletion struct {
	Shell ShellType `arg:"" help:"Shell type to generate completion script for."`
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

	fmt.Println(complete)
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
