package main

import (
	"github.com/sirkon/errors"
	"github.com/sirkon/fenneg"
	"github.com/sirkon/message"
)

func main() {
	handlers, err := fenneg.NewTypesHandlers()
	if err != nil {
		message.Fatal(errors.Wrap(err, "that is truly unexpected"))
	}

	r, err := fenneg.NewRunner("github.com/sirkon/errors", handlers)
	if err != nil {
		message.Fatal(errors.Wrap(err, "create fenneg runner"))
	}

	const pkg = "github.com/sirkon/cmd-tools/cmd/portalctl/internal/portallog"
	err = r.
		OpLog().
		Source(pkg, "PortalLogger").
		Type(pkg, "Encoding").
		LengthPrefix(true).
		Run()
	if err != nil {
		message.Fatal(errors.Wrap(err, "generate type for portal log records"))
	}
}
