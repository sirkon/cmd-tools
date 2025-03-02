package main

import (
	"github.com/sirkon/cmd-tools/cmd/portalctl/internal/portallog"
	"github.com/sirkon/errors"
)

// CommandLogCompact реализация операции log-compact.
type CommandLogCompact struct{}

// Run запуск команды.
func (d CommandLogCompact) Run(ctx *RunContext) error {
	data, err := portallog.LogRead(ctx.opLogFile)
	if err != nil {
		return errors.Wrap(err, "read op log file data")
	}

	if err := portallog.LogDump(ctx.opLogFile, ctx.appCacheRoot, data); err != nil {
		return errors.Wrap(err, "dump current op log data")
	}

	return nil
}
