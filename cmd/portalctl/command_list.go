package main

import (
	"os"
	"sort"
	"strings"

	"github.com/sirkon/cmd-tools/cmd/portalctl/internal/portallog"
	"github.com/sirkon/errors"
)

// CommandList реализация команды list.
type CommandList struct {
	Prefix string `arg:"" help:"Prefix filter." default:""`
}

// Run запуск команды.
func (d CommandList) Run(ctx *RunContext) error {
	data, err := portallog.LogRead(ctx.opLogFile)
	if err != nil {
		return errors.Wrap(err, "read op log data")
	}

	var names []string
	var width int
	for name := range data {
		if !strings.HasPrefix(name, d.Prefix) {
			continue
		}

		if len(name) > width {
			width = len(name)
		}

		names = append(names, name)
	}

	curdir, _ := os.Getwd()

	sort.Strings(names)
	var buf strings.Builder
	for _, name := range names {
		buf.Reset()

		path := data[name]
		current := path == curdir
		if current {
			buf.WriteString("\033[1;32m")
		}
		buf.WriteString(name)
		buf.WriteString(strings.Repeat(" ", width-len(name)+1))
		buf.WriteString("-> ")
		buf.WriteString(path)
		if current {
			buf.WriteString("\033[0m")
		}
		buf.WriteByte('\n')
		_, _ = os.Stdout.WriteString(buf.String())
	}

	return nil
}
