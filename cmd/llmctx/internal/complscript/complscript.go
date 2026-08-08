package complscript

import (
	_ "embed"
)

// Fish is the comptime skeleton completion for the fish shell. Bash and Zsh
// completions are generated on the fly from the live kong model (king has no
// positional completion support for fish), so only the fish skeleton is kept
// embedded here.
//
//go:embed fish
var Fish string
