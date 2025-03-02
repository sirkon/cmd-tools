package main

import (
	"os"

	"github.com/sirkon/cmd-tools/cmd/portalctl/internal/portallog"
	"github.com/sirkon/errors"
)

// CommandHere реализация команды here.
type CommandHere struct {
	Name PortalName `arg:"" required:"true" help:"Portal name."`
}

// Run запуск команды.
func (d CommandHere) Run(ctx *RunContext) error {
	data, err := portallog.LogRead(ctx.opLogFile)
	if err != nil {
		return errors.Wrap(err, "read op log data")
	}

	if _, ok := data[string(d.Name)]; ok {
		return errors.Wrapf(err, "cannot overwrite existing portal %q", d.Name)
	}

	curdir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "get current directory path")
	}

	oplog := portallog.NewLog(ctx.opLogFile)
	if err := oplog.AddPortal(string(d.Name), curdir); err != nil {
		return errors.Wrap(err, "store data")
	}

	return nil
}
