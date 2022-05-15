# jishell - A powerful modern CLI and SHELL,with a msfconsole-like style

[![GoDoc](https://godoc.org/github.com/chroblert/jishell?status.svg)](https://godoc.org/github.com/chroblert/jishell)
[![Go Report Card](https://goreportcard.com/badge/github.com/chroblert/jishell)](https://goreportcard.com/report/github.com/chroblert/jishell)

本项目基于grumble进行开发，在其基础上做了如下改变：

- list类型的`arg`和`flag`，基于`,`进行分割
- 子命令中可使用多个list类型的`arg`
- 使用具有`msfconsole`风格的shell
- 增加`setf`,`seta`,`unsetf`,`unseta`,`help`,`run`,`show`，`use`等核心命令，命令说明如下：
  - `setf`:用来设置`flag`
  - `seta`:用来设置`arg`
  - `unsetf`:取消设置的`flag`
  - `unseta`:取消设置的`arg`
  - `help`:查看特定命令的说明
  - `show`:显示当前命令的`arg`和`flag`值
  - `run`:执行当前命令
  - `use`:切换命令
- 修改词法分析器`shlex`，使之能够完成基于特定字符的分割
- 为`seta`,`setf`,`unseta`,`unsetf`增加自动补全
- 支持三种模式
  - 控制台模式：`./samples cdnChkD --boolf test.com false`
  - 普通交互模式：
    - `./samples`进入交互模式
    - `cdnChkD --boolf test.com false`
  - `msfconsole`风格模式：
    - `./samples`进入交互模式
    - 使用`use`切换到目标子命令
    - 使用`seta`,`setf`设置`flag`和`arg`
    - 使用`run`执行命令
- ...

![](https://gitee.com/chroblert/pictures/raw/master/img/20220515210929.png)


## Introduction

目录层级

```shell
|main.go 					// 启动函数
|cmd     
|-app.go  					// APP
|-CDNCheck
|---------cdn_chk_ip.go		// subCommand
|---------cdn_chk_domain.go // subCommand

```

![](https://gitee.com/chroblert/pictures/raw/master/img/20220515204313.png)

`app.go`创建一个`jishell APP`

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

`app.go`加载存储子命令

```go
func init() {
	App.SetPrintASCIILogo(func(a *jishell.App) {
		jlog.Warn("=============================")
		jlog.Warn("         CDN Check          ")
		jlog.Warn("=============================")
	})
	App.OnInit(func(a *jishell.App, flags jishell.FlagMap) error {
		if flags.Bool("verbose") {
			jlog.Info("verbose")
		}
		// load commands
		if viper.Get("jCommands") != nil {
			for _, v := range viper.Get("jCommands").([]*jishell.Command) {
				a.AddCommand(v)
			}
		}
		return nil
	})
}
```

`cdn_chk_domain.go`创建一个APP下的子命令

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

`cdn_chk_domain.go`存储子命令

```go
func init() {
	var tmpCommands []*jishell.Command
	if viper.Get("jCommands") == nil {
		tmpCommands = make([]*jishell.Command, 0)
	} else {
		tmpCommands = viper.Get("jCommands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, cdnChkDomainCmd)
	viper.Set("jCommands", tmpCommands)
}
```

`main.go`运行

```go
err := App.Run()
```

Or

使用内置函数`*jishell.Main*` 自动处理异常。

```go
func main() {
	jishell.Main(App)
}
```

## 示例

[https://github.com/chroblert/jishell/tree/master/samples](https://github.com/chroblert/jishell/tree/master/samples)

## Credits

This project is based on ideas from the great [grumble](https://github.com/desertbit/grumble) library.

## License

MIT
