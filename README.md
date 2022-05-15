# jishell - A powerful modern CLI and SHELL,with a msfconsole-like style

[![GoDoc](https://godoc.org/github.com/chroblert/jishell?status.svg)](https://godoc.org/github.com/chroblert/jishell)
[![Go Report Card](https://goreportcard.com/badge/github.com/chroblert/jishell)](https://goreportcard.com/report/github.com/chroblert/jishell)


## Introduction

目录层级

创建一个`jishell APP`

```go
var App = jishell.New(&jishell.Config{
	Name:                  "CheckTest",
	Description:           "",
	Flags: func(f *jishell.Flags) {
		f.BoolL("verbose",false,"")
	},
	CurrentCommand:        "CheckTest", // shell中使用，一定不要与其他子命令的Name值重复。建议与该app Name值相同。
})
```

创建一个APP下的子命令

```go
var cdnChkDomainCmd = &jishell.Command{
	Name:      "cdnChkD",
	Help:      "检测目标域名是否使用了CDN",
	LongHelp:  "",
	HelpGroup: "CDN Check",
	Usage:     "",
	Flags: func(f *jishell.Flags) {
		f.Bool("b","boolf",true,"")
	},
	Args: func(a *jishell.Args) {
		a.String("host", "hostname.eg: www.test.com")
		a.Bool("verbose","")
        a.StringList("t", "")
	},
	Run: func(c *jishell.Context) error {
		jlog.Info("boolf:",c.Flags.Bool("boolf"))
		jlog.Info("host:",c.Args.String("host"))
		jlog.Info("verbose:",c.Args.Bool("verbose"))
        jlog.Info("t:",c.Args.StringList("t"))
		return nil
	},
	CMDPath:   "Info/CDN", // 用来标记该命令所在的分组。用于use的自动补全
}
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
