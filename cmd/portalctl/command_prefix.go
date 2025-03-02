package main

import (
	"fmt"

	"github.com/sirkon/cmd-tools/cmd/portalctl/internal/portallog"
)

// CommandPrefix реализация команды prefix.
type CommandPrefix struct {
	Prefix string `arg:"" help:"Portal name prefix." default:""`
}

// Run запуск команды.
func (d CommandPrefix) Run(ctx *RunContext) error {
	portals, err := portallog.LogFilter(ctx.opLogFile, string(d.Prefix))
	if err != nil {
		return nil
	}

	for _, portal := range portals {
		fmt.Println(portal)
	}

	return nil
}
