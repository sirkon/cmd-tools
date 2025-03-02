package main

import (
	"runtime/debug"

	"github.com/sirkon/message"
)

// CommandVersion реализация команды version.
type CommandVersion struct{}

// Run запуск команды.
func (CommandVersion) Run(ctx *RunContext) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		message.Warning(
			"WARNING: you are using a version compiled with modules disabled, this is not the way it supposed to be.",
		)
	} else {
		message.Info(appName, "version", info.Main.Version)
	}

	return nil
}
