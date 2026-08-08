# Move to commands for bash/zsh/fish completion from ShellType values.

You need change from current @cmd/llcmt/command_completion.go approach with

```go
type CommandCompletion struct {
	Shell ShellType `arg:""    help:"Shell type to generate completion script for."`
	Save  bool      `short:"s" help:"Save completion script for llmctx within completion files."`
}
```

to 

```go
type CommandCompletion struct {
    Bash   CommandCompletionBash `cmd:"" help:"Generate completion script for Bash."`
	// ...
	
    Save  bool      `short:"s" help:"Save completion script for llmctx within completion files."`
}

type CommandCompletionBash struct {}

type CommandCompletionZsh struct {}

type CommandCompletionFish struct {}

// ...
```

## What will change.

- No more `Run` method for `CommandCompletion`. Write `(CommandCompletion) completionBash() error` instead and so on
  for `Zsh`, etc. Remove `Run` from `CommandCompletion`.
- `CommandCompletionBash` and such are getting `Run` method as follows
  ```go
  func (CommandCompletionBash) Run(ctx *runContext) error {
      return ctx.cli.Completion.completionBash()
  }
  ```
- Implement those `completionBash`, `completionZsh` and `completionFish`. They should follow the current semantics
  and either output script into the stdout or write it into places when `Save` is `true`.

## What's the point.

Self-documentation. Shell types values were not represented in kong help output. They will with separated commands.



