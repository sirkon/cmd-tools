package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

var (
	rcappend = []string{
		"portal() {",
		"    cd `portalctl show $1`",
		"}",
		"complete -o nospace -C 'portalctl prefix' portal",
	}
)

// CommandSetup реализация команды setup.
type CommandSetup struct{}

// Run запуск команды.
func (d CommandSetup) Run(ctx *RunContext) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get user homedir")
	}

	filesToTouch := []string{
		".bashrc",
		".zshrc",
	}
	for _, file := range filesToTouch {
		if err := installIntoFile(homedir, file); err != nil {
			if !errors.Is(err, justPassThisRC) {
				return errors.Wrap(err, "process "+file)
			}
			continue
		}

		message.Infof("%s done", file)
	}

	return nil
}

const justPassThisRC = errors.Const("this file was not found and it is OK")

func installIntoFile(home, rc string) error {
	rcFullPath := filepath.Join(home, rc)
	stat, err := os.Stat(rcFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			message.Infof("%s not found, omitting", rc)
			return justPassThisRC
		}
	}
	if !stat.Mode().IsRegular() {
		return errors.New("is not a regular file")
	}

	data, err := os.ReadFile(rcFullPath)
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, rcappend[0]) {
			message.Infof("%s has already been set up, omitting", rc)
			return nil
		}
	}

	// Убираем пустые строки в конце
	for len(lines) > 0 {
		if lines[len(lines)-1] != "" {
			break
		}

		lines = lines[:len(lines)-1]
	}

	var builder bytes.Buffer
	for _, line := range lines {
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	builder.WriteByte('\n')
	for _, line := range rcappend {
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	builder.WriteByte('\n')

	rcTmpPath := filepath.Join(home, ".rc-tmp")
	if err := os.WriteFile(rcTmpPath, builder.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "build a temporary file with prerequisites")
	}

	if err := os.Rename(rcTmpPath, rcFullPath); err != nil {
		return errors.Wrap(err, "replace original with the temporary file")
	}

	return nil
}
