package main

import (
	"os"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

// CommandDelete удалить существующий контекст из локальной базы.
type CommandDelete struct {
	Name string `arg:"" required:"" help:"Context name to delete."`
}

// Run запуск команды.
func (c *CommandDelete) Run(ctx *runContext) error {
	viewer, err := ctx.interpreter()
	if err != nil {
		return errors.Wrap(err, "interpret log data")
	}

	info := viewer.Info(c.Name)
	if info == nil {
		return errors.New("unknown context " + c.Name)
	}

	// Вначале помечаем запись удаленной.
	if err := ctx.deleter(c.Name); err != nil {
		return errors.Wrap(err, "write context delete operation")
	}

	// Пытаемся удалить файл с данными контекста.
	if err := os.Remove(info.Path); err != nil {
		message.Warningf("failed to remove %q, you may want to delete it manually", info.Path)
		return errors.Wrap(err, "delete context data file")
	}

	return nil
}
