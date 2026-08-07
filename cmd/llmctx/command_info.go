package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

// CommandInfo показать данные текущего контекста.
type CommandInfo struct {
	Name string `arg:"" required:"" help:"Context name to put into the current git project."`
}

func (c *CommandInfo) Run(ctx *runContext) error {
	viewer, err := ctx.interpreter()
	if err != nil {
		return errors.Wrap(err, "interpret log data")
	}

	info := viewer.Info(c.Name)
	if info == nil {
		return errors.New("unknown context " + c.Name)
	}

	f, err := os.Open(info.Path)
	if err != nil {
		return errors.Wrap(err, "read context data file")
	}
	defer func() {
		if err := f.Close(); err != nil {
			message.Warningf("failed to close file %q: %v", info.Path, err)
		}
	}()

	fmt.Printf("name: %s\n", info.Name)
	fmt.Printf("description: %s\n", info.Desc)
	fmt.Println("===========================")
	if _, err := io.Copy(os.Stdout, f); err != nil {
		return errors.Wrap(err, "copy context data file")
	}

	return nil
}
