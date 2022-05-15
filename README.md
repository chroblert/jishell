# jishell - A powerful modern CLI and SHELL,with a msfconsole-like style

[![GoDoc](https://godoc.org/github.com/chroblert/jishell?status.svg)](https://godoc.org/github.com/chroblert/jishell)
[![Go Report Card](https://goreportcard.com/badge/github.com/chroblert/jishell)](https://goreportcard.com/report/github.com/chroblert/jishell)


## Introduction

Create a jishell APP.

```go
var app = jishell.New(&jishell.Config{
	Name:        "app",
	Description: "short app description",

	Flags: func(f *jishell.Flags) {
		f.String("d", "directory", "DEFAULT", "set an alternative directory path")
		f.Bool("v", "verbose", false, "enable verbose mode")
	},
})
```

Register a top-level command. *Note: Sub commands are also supported...*

```go
app.AddCommand(&jishell.Command{
    Name:      "daemon",
    Help:      "run the daemon",
    Aliases:   []string{"run"},

    Flags: func(f *jishell.Flags) {
        f.Duration("t", "timeout", time.Second, "timeout duration")
    },

    Args: func(a *jishell.Args) {
        a.String("service", "which service to start", jishell.Default("server"))
    },

    Run: func(c *jishell.Context) error {
        // Parent Flags.
        c.App.Println("directory:", c.Flags.String("directory"))
        c.App.Println("verbose:", c.Flags.Bool("verbose"))
        // Flags.
        c.App.Println("timeout:", c.Flags.Duration("timeout"))
        // Args.
        c.App.Println("service:", c.Args.String("service"))
        return nil
    },
})
```
```
List类型的值以,进行分隔。
控制台中输入List类型的值，建议以'进行包裹。
- 若想输入',"，则需要使用\进行转义。如：'1,\",2 经过处理后，List包含1,"和2两个元素
shell中没有限制.
```

Run the application.

```go
err := app.Run()
```

Or use the builtin *jishell.Main* function to handle errors automatically.

```go
func main() {
	jishell.Main(app)
}
```

## Shell Multiline Input

Builtin support for multiple lines.

```
>>> This is \
... a multi line \
... command
```

## Samples



## Additional Useful Packages

- https://github.com/AlecAivazis/survey
- https://github.com/tj/go-spin

## Credits

This project is based on ideas from the great [grumble](https://github.com/desertbit/grumble) library.

## License

MIT
