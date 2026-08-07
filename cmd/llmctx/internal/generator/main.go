package main

import (
	"github.com/sirkon/errors"
	"github.com/sirkon/fenneg"
	"github.com/sirkon/message"
)

const metadataPkg = "github.com/sirkon/cmd-tools/cmd/llmctx/internal/metadata"

func main() {
	errors.InsertLocations()
	message.SetVerbosityLevelFullContext()

	hnlrs, err := fenneg.NewTypesHandlers()
	if err != nil {
		message.Critical(errors.Wrap(err, "set up type handlers"))
	}

	r, err := fenneg.NewRunner("github.com/sirkon/errors", hnlrs)
	if err != nil {
		message.Fatal(errors.Wrap(err, "setup codegen runner"))
	}

	if err := r.OpLog().
		Source(metadataPkg, "Operations").
		Type(metadataPkg, "OperationsLogger").
		LengthPrefix(true).
		Run(); err != nil {
		message.Critical(errors.Wrap(err, "process operations"))
	}

	if err := r.OpLog().
		Source(metadataPkg, "Dump").
		Type(metadataPkg, "DumpLogger").
		LengthPrefix(true).
		Run(); err != nil {
		message.Critical(errors.Wrap(err, "process dump"))
	}
}
