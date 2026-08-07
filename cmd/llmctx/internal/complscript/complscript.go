package complscript

import (
	_ "embed"
)

//go:embed bash
var Bash string

//go:embed zsh
var Zsh string

//go:embed fish
var Fish string
