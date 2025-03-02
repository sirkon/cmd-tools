package main

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

const (
	appName = "portalctl"
)

func main() {
	cachesRoot, err := os.UserCacheDir()
	if err != nil {
		message.Fatal(errors.Wrap(err, "get user cache directory path"))
	}

	appCacheDir := filepath.Join(cachesRoot, appName)
	if err := os.MkdirAll(appCacheDir, 0755); err != nil {
		message.Fatal(errors.Wrap(err, "create cache dir for the app"))
	}

	var cmd Commands

	parser := kong.Must(
		&cmd,
		kong.Name(appName),
		kong.Description(
			"Portal control utility.",
		),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.UsageOnError(),
		kong.Vars{},
	)
	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	err = ctx.Run(&RunContext{
		cmd:          &cmd,
		appCacheRoot: appCacheDir,
		opLogFile:    filepath.Join(appCacheDir, appName+".bin"),
	})
	if err != nil {
		message.Fatal(errors.Wrap(err, "run command"))
	}
}
