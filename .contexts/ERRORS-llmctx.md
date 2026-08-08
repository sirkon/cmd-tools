# Error processing here.

You may see something like

```go
message.Warning(errors.Wrap(err, "failed to open file"))
```

throughout the code and which looks like a violation of "do not say it is an error, say what were you doing".
But it is not. Here `message.Warning` (`message.Debug`, `message.Info`, etc) are loggers from
`github.com/sirkon/message`. And the "annotation" is actually a logging text. 