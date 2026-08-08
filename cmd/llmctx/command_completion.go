package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/miekg/king"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

// CommandCompletion команда генерации скриптов автодополнения.
type CommandCompletion struct {
	Bash CommandCompletionBash `cmd:"" help:"Generate completion script for Bash."`
	Zsh  CommandCompletionZsh  `cmd:"" help:"Generate completion script for Zsh."`
	Fish CommandCompletionFish `cmd:"" help:"Generate completion script for Fish."`

	Save bool `short:"s" help:"Save completion script for llmctx within completion files."`
}

// CommandCompletionBash команда генерации скрипта автодополнения для Bash.
type CommandCompletionBash struct{}

// CommandCompletionZsh команда генерации скрипта автодополнения для Zsh.
type CommandCompletionZsh struct{}

// CommandCompletionFish команда генерации скрипта автодополнения для Fish.
type CommandCompletionFish struct{}

func (CommandCompletionBash) Run(ctx *runContext) error {
	return ctx.cli.Completion.completionBash()
}

func (CommandCompletionZsh) Run(ctx *runContext) error {
	return ctx.cli.Completion.completionZsh()
}

func (CommandCompletionFish) Run(ctx *runContext) error {
	return ctx.cli.Completion.completionFish()
}

// completionBash генерирует скрипт автодополнения для Bash.
func (c *CommandCompletion) completionBash() error {
	complete, err := c.generate(&king.Bash{}, contextNamesHelper)
	if err != nil {
		return errors.Wrap(err, "generate Bash completion script")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get current user home directory")
	}
	absComplFileName := filepath.Join(home, ".local", "share", "bash-completion", "completions", appName)

	return c.output(complete, absComplFileName)
}

// completionZsh генерирует скрипт автодополнения для Zsh.
func (c *CommandCompletion) completionZsh() error {
	complete, err := c.generate(&king.Zsh{}, contextNamesHelperZsh)
	if err != nil {
		return errors.Wrap(err, "generate Zsh completion script")
	}

	// king оборачивает позиционные аргументы в `_values "name" $(CMD)`, что
	// отбрасывает описания. Заменяем это на прямой вызов функции-экшена,
	// который использует _describe с парами name:description.
	complete = strings.ReplaceAll(complete, `_values "name" $(_llmctx_context_names)`, `_llmctx_context_names`)

	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get current user home directory")
	}
	absComplFileName := filepath.Join(home, ".zsh", "completions", "_"+appName)

	return c.output(complete, absComplFileName)
}

// completionFish возвращает скелет автодополнения для Fish. У king нет
// автодополнения позиционных аргументов для Fish, поэтому используется
// вшитый скелет из complscript.
func (c *CommandCompletion) completionFish() error {
	complete, err := c.generate(&king.Fish{}, contextNamesHelper)
	if err != nil {
		return errors.Wrap(err, "generate Bash completion script")
	}

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return errors.Wrap(err, "get current user config directory")
	}
	absComplFileName := filepath.Join(cfgDir, "fish", "completions", appName+".fish")

	return c.output(complete, absComplFileName)
}

// generate рендерит скрипт автодополнения через king. Скрипт разрешает имена
// контекстов в рантайме благодаря тегам `completion:"_llmctx_context_names"` и
// добавленному ниже хелперу.
func (CommandCompletion) generate(completer king.Completer, helper string) (string, error) {
	cmd, err := kong.New(&CLI{})
	if err != nil {
		return "", errors.Wrap(err, "build kong model for completion generation")
	}
	completer.Completion(cmd.Model.Node, appName)
	return string(completer.Out()) + "\n\n" + helper, nil
}

// output печатает скрипт в stdout, либо сохраняет его в файлы автодополнения,
// когда установлен Save.
func (c *CommandCompletion) output(complete string, absComplFileName string) error {
	if !c.Save {
		_, _ = fmt.Fprintln(os.Stdout, complete)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(absComplFileName), 0755); err != nil {
		return errors.Wrap(err, "create completion directory")
	}
	if err := os.WriteFile(absComplFileName, []byte(complete), 0644); err != nil {
		return errors.Wrap(err, "write completion script")
	}

	message.Infof("Completion script saved to %s\n", absComplFileName)

	return nil
}

// contextNamesHelper reads `llmctx list --porcelain` (NULL-terminated
// name\0desc\0 pairs) and prints the context names one per line. Bash and zsh
// both word-split the output of the helper, so the same function serves both.
const contextNamesHelper = `_llmctx_context_names() {
  # Context names from 'llmctx list --porcelain' (name\0desc\0...).
  local name desc
  while read -r -d '' name && read -r -d '' desc; do
    printf '%s\n' "$name"
  done < <(llmctx list --porcelain)
}
`

// contextNamesHelperZsh выводит контексты в виде пар name:description для zsh.
// Используется как экшен автодополнения через _describe, поэтому zsh показывает
// описания рядом с именами.
const contextNamesHelperZsh = `_llmctx_context_names() {
  # Context names from 'llmctx list --porcelain' (name\0desc\0...).
  local -a items
  local name desc
  while IFS= read -r -d '' name && IFS= read -r -d '' desc; do
    items+=("$name:$desc")
  done < <(llmctx list --porcelain)
  _describe 'context' items
}
`
