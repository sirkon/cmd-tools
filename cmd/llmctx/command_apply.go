package main

import (
	"os"
	"path/filepath"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

// CommandApply сохранить существующий контекст в папке контекстов текущего репозитория git.
type CommandApply struct {
	Name string `arg:"" required:"" help:"Context name to put into the current git project."`
	File string `arg:"" required:"" help:"File name to store context into."`
}

func (c *CommandApply) Run(ctx *runContext) (err error) {
	var dstPath string

	gitRoot, err := getGitRoot()
	if err != nil {
		message.Warning(errors.Wrap(err, "failed to get git root"))
		message.Info("will save context right into the current folder")
		dstPath = c.File
	} else {
		contextsDir := filepath.Join(gitRoot, ".contexts")
		// Создаем папку .contexts, если её ещё нет (если есть — ничего не сделает)
		if err := os.MkdirAll(contextsDir, 0755); err != nil {
			return errors.Wrap(err, "failed to create .contexts directory")
		}
		dstPath = filepath.Join(contextsDir, c.File)
	}

	viewer, err := ctx.interpreter()
	if err != nil {
		return errors.Wrap(err, "interpret log data")
	}

	info := viewer.Info(c.Name)
	if info == nil {
		return errors.New("unknown context " + c.Name)
	}

	if err := copyFile(info.Path, dstPath); err != nil {
		return errors.Wrap(err, "copy context data into storage")
	}

	return nil
}
