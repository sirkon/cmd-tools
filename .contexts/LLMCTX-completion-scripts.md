# Completion scripts of llmctx.

The best autocompletion solution for github.com/alecthomas/kong is using github.com/miekg/king.

But it comes with shortcomings: the output it produces is fixed and cannot autocomplete things not known 
during the compile time.

Currently, we override this by a two-step solution:

1. Generate completion script by github.com/miekg/king completer:
   ```go
   cmd, _ := kong.New(&CLI{})
   completer := &king.Bash{} // &king.Zsh{}, &king.Fish{}
   completer.Completion(cmd.Model.Node, appName)
   fmt.Println(string(completer.Out()))
   ```
   which will generate autocompletion for the shell type chosen.
2. Tweak it manually to handle runtime knowledge.

Here in llmctx we autocomplete names in `llmctx apply`, `llmctx info`, etc by using `llmctx list --porcelain` output.
Which is a NULL-terminated sequence of | name1 | description1 | name2 | ...

When command is updated we use @cmd/llmct/command_completion_test.go run to get the king output and apply
runtime resolution.

---
Alternatively, we may handle runtime logic with the use of `completion:"xxx"` tag on fields, but I don't know the details.
In such case, we can get rid of @cmd/llmctx/complscript altogether and just rely on direct completer.Output()
at runtime.

Please discover how the completion may be used and will it solve the problem.