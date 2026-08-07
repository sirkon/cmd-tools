package main

import (
	"os"
)

func editorName() string {
	name := os.Getenv("LLMCTX_EDITOR")
	if name != "" {
		return name
	}

	name = os.Getenv("EDITOR")
	if name != "" {
		return name
	}

	return ""
}

func getCommonEditorPreset(name string) *editorPreset {
	v, ok := knownEditors[name]
	if ok {
		return &v
	}

	return nil
}

var knownEditors = map[string]editorPreset{
	"zed":  {Binary: "zed", Flags: []string{"--wait"}},
	"code": {Binary: "code", Flags: []string{"--wait"}},
	"subl": {Binary: "subl", Flags: []string{"-w"}},
	"atom": {Binary: "atom", Flags: []string{"--wait"}},
	"mate": {Binary: "mate", Flags: []string{"-w"}},
	"vim":  {Binary: "vim", Flags: []string{}},
	"nvim": {Binary: "nvim", Flags: []string{}},
	"nano": {Binary: "nano", Flags: []string{}},
}

type editorPreset struct {
	Binary string   // Имя бинарника в системе
	Flags  []string // Флаги, обязательные для CLI (например, для ожидания закрытия)
}
