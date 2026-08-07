package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/sirkon/errors"
)

// CommandStash сохранить файл как контекст с данным именем и описанием.
type CommandStash struct {
	Name        ValidName        `short:"n" required:"" help:"Name to stash context under."`
	Description ValidDescription `short:"d" required:"" help:"Description of context."`
	File        string           `arg:""    required:"" help:"Context file."`
}

func (c *CommandStash) Run(ctx *runContext) error {
	viewer, err := ctx.interpreter()
	if err != nil {
		return errors.Wrap(err, "interpret log")
	}

	if viewer.Info(string(c.Name)) != nil {
		return errors.New("there's already a stashed context with this name")
	}

	path := filepath.Join(ctx.cacheRoot, string(c.Name)+"-"+strconv.FormatInt(time.Now().Unix(), 10))
	if err := copyFile(c.File, path); err != nil {
		return errors.Wrap(err, "copy context data")
	}

	if err := ctx.inserter(ContextData{
		Name: string(c.Name),
		Desc: string(c.Description),
		Path: path,
	}); err != nil {
		return errors.Wrap(err, "write context metadata")
	}

	return nil
}

type ValidName string

func (v *ValidName) UnmarshalText(raw []byte) error {
	if !nameValidMask.Match(raw) {
		return fmt.Errorf("invalid name %q, must be alpha[-alnum]*, like some-name, some-name-2, etc", string(raw))
	}

	*v = ValidName(raw)
	return nil
}

var nameValidMask = regexp.MustCompile(`^[a-zA-Z]+(?:-[a-zA-Z0-9]+)*$`)

type ValidDescription string

func (v *ValidDescription) UnmarshalText(raw []byte) error {
	if len(raw) == 0 {
		return errors.New("empty strings are not allowed")
	}

	*v = ValidDescription(raw)
	return nil
}
