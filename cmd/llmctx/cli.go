package main

import (
	"iter"
	"maps"
	"slices"
)

// CLI структура команды.
type CLI struct {
	Apply  CommandApply  `cmd:"" help:"Add given context into the current project."`
	Delete CommandDelete `cmd:"" help:"Delete given context."`
	Edit   CommandEdit   `cmd:"" help:"Edit text of the given context."`
	Info   CommandInfo   `cmd:"" help:"Show context information."`
	List   CommandList   `cmd:"" help:"List all contexts."`
	Stash  CommandStash  `cmd:"" help:"Stash context."`

	Completion CommandCompletion `cmd:"" help:"Generate completion code."`
}

type runContext struct {
	fileEdit    func(name string) error
	interpreter func() (*MetadataViewer, error)
	inserter    func(data ContextData) error
	deleter     func(name string) error

	cacheRoot string
}

// MetadataViewer сущность для показа метаданных.
type MetadataViewer struct {
	ops map[string]ContextData
}

// Add не вызывать!
func (o *MetadataViewer) Add(name string, description string, path string) error {
	o.ops[name] = ContextData{
		Name: name,
		Desc: description,
		Path: path,
	}

	return nil
}

// Del не вызывать!
func (o *MetadataViewer) Del(name string) error {
	delete(o.ops, name)

	return nil
}

// Info информация об контексте.
func (o *MetadataViewer) Info(name string) *ContextData {
	res, ok := o.ops[name]
	if !ok {
		return nil
	}

	return &res
}

// Iter итератор по контексту.
func (o *MetadataViewer) Iter() iter.Seq[ContextData] {
	return func(yield func(ContextData) bool) {
		keys := slices.Collect(maps.Keys(o.ops))
		slices.Sort(keys)

		for _, key := range keys {
			if !yield(o.ops[key]) {
				return
			}
		}
	}
}

// ContextData данные контекста.
type ContextData struct {
	Name string
	Desc string
	Path string
}
