package main

import (
	"fmt"

	"github.com/sirkon/cmd-tools/cmd/portalctl/internal/portallog"
	"github.com/sirkon/errors"
)

// CommandShow реализация команды show.
type CommandShow struct {
	Name PortalName `arg:"" required:"true" help:"Portal name."`
}

// Run запуск команды.
func (d CommandShow) Run(ctx *RunContext) error {
	path, err := portallog.LogShowPortalPath(ctx.opLogFile, string(d.Name))
	if err != nil {
		return errors.Wrapf(err, "look for portal %q path", d.Name)
	}

	fmt.Println(path)
	return nil
}
