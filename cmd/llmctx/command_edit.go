package main

import (
	"os"

	"github.com/sirkon/errors"
)

// CommandEdit команда для редактирования текста существующего контекста.
type CommandEdit struct {
	Name string `arg:"" required:"" completion:"_llmctx_context_names" help:"Context name to edit."`
}

func (c *CommandEdit) Run(ctx *runContext) error {
	viewer, err := ctx.interpreter()
	if err != nil {
		return errors.Wrap(err, "interpret log data")
	}

	info := viewer.Info(c.Name)
	if info == nil {
		return errors.New("unknown context " + c.Name)
	}

	if err := ctx.fileEdit(info.Path); err != nil {
		return errors.Wrap(err, "edit context data file")
	}

	// Проверяем, существует ли файл на диске после работы редактора.
	if _, err := os.Stat(info.Path); err != nil {
		return errors.Wrapf(err, "context data file %q is missing after editing", info.Path)
	}

	return nil
}
