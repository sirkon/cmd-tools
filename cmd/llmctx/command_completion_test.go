package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/miekg/king"

	"github.com/sirkon/cmd-tools/cmd/llmctx/internal/complscript"
)

func TestCompletionScriptGen(t *testing.T) {
	cases := []struct {
		shell   string
		compile king.Completer
		helper  string
		runtime bool // разрешает ли скрипт имена контекстов в рантайме
	}{
		{shell: "bash", compile: &king.Bash{}, helper: contextNamesHelper, runtime: true},
		{shell: "zsh", compile: &king.Zsh{}, helper: contextNamesHelperZsh, runtime: true},
		{shell: "fish", runtime: false},
	}

	output := make(map[string]string)
	for _, c := range cases {
		var script string
		var err error
		if c.compile == nil {
			script = complscript.Fish
		} else {
			var gen CommandCompletion
			script, err = gen.generate(c.compile, c.helper)
			if err != nil {
				t.Fatalf("generate(%s): %v", c.shell, err)
			}
		}

		if c.runtime != strings.Contains(script, "_llmctx_context_names") {
			t.Errorf("shell %s: наличие хелпера runtime (%v) не совпадает с ожидаемым (%v)",
				c.shell, strings.Contains(script, "_llmctx_context_names"), c.runtime)
		}
		if c.compile != nil {
			output[c.shell] = script
		}
	}

	result, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("не удалось преобразовать вывод в json: %v", err)
	}

	if _, err := t.Output().Write(result); err != nil {
		t.Fatalf("не удалось записать вывод: %v", err)
	}
}
