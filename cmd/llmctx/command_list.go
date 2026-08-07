package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirkon/errors"
)

// CommandList показать список существующих контекстов.
type CommandList struct {
	Porcelain bool `help:"Machine friendly NULL-terminated output"`
}

func (c *CommandList) Run(ctx *runContext) error {
	viewer, err := ctx.interpreter()
	if err != nil {
		return errors.Wrap(err, "interpret log data")
	}

	var nullTerminator [1]byte
	if c.Porcelain {
		for data := range viewer.Iter() {
			_, _ = io.WriteString(os.Stdout, data.Name)
			_, _ = os.Stdout.Write(nullTerminator[:])
			_, _ = io.WriteString(os.Stdout, data.Desc)
			_, _ = os.Stdout.Write(nullTerminator[:])
		}
		return nil
	}

	width := 0
	for data := range viewer.Iter() {
		if len(data.Name) > width {
			width = len(data.Name)
		}
	}

	fmt.Printf("%-*s  description\n", width, "name")
	fmt.Println(strings.Repeat("=", width+2+len("description")))

	for data := range viewer.Iter() {
		fmt.Printf("%-*s  %s\n", width, data.Name, data.Desc)
	}

	return nil
}
