package jishell

import (
	"fmt"
	"github.com/chroblert/go-shlex"
	"github.com/chroblert/jishell/jconfig"
	"github.com/desertbit/readline"
	"github.com/jedib0t/go-pretty/v6/table"
	"reflect"
	"strings"
)

func core_help(a *App) *Command {
	return &Command{
		Name: "help",

		Help:      "use 'help [command]' for command help",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Args: func(a *Args) {
			a.StringList("command", "the name of the command")
		},
		Run: func(c *Context) error {
			args := c.Args.StringList("command")
			if len(args) == 0 {
				if c.App.currentCmd == nil {
					a.printHelp(a, a.isShell)
				} else {
					a.printCommandHelp(a, c.App.currentCmd, a.isShell, len(args) > 0)
				}
				return nil
			}
			var cmd *Command
			var err error
			if c.App.currentCmd == nil {
				cmd, _, err = a.commands.FindCommand(args)
			} else {
				cmd, _, err = a.currentCmd.commands.FindCommand(args)
			}
			if err != nil {
				return err
			} else if cmd == nil {
				a.PrintError(fmt.Errorf("command not found"))
				return nil
			}
			a.printCommandHelp(a, cmd, a.isShell, len(args) > 0)
			return nil
		},
		isBuiltin: true,
	}
}

func core_exit(a *App) *Command {
	return &Command{
		Name: "exit",

		Help:      "exit the shell",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Run: func(c *Context) error {
			c.Stop()
			return nil
		},
		isBuiltin: true,
	}
}

func core_clear(a *App) *Command {
	return &Command{
		Name: "clear",

		Help:      "clear the screen",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Run: func(c *Context) error {
			readline.ClearScreen(a.rl)
			return nil
		},
		isBuiltin: true,
	}
}

func core_use(a *App) *Command {
	return &Command{
		Name: "use",

		Aliases:   nil,
		Help:      "switch command",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "use <command|Alias|CMDPath>",
		//Flags:     nil,
		Args: func(a *Args) {
			a.String("commandName", "command name")
		},
		Run: func(c *Context) error {
			// 获取要切换的command
			inputCmdStr := c.Args.String("commandName")
			//tmpStrSlice := strings.Split(c.Args.String("commandName"),"/")
			//commandName := tmpStrSlice[len(tmpStrSlice)-1]
			//commandCategory := tmpStrSlice[0]
			var tmpCommand = &Command{}
			// JC 220520 判断是否处在app本身
			if c.App.currentCmd == nil {
				tmpCommand = c.App.Commands().Get(inputCmdStr)
			} else {
				// 否则取回当前命令的子命令
				tmpCommand = c.App.currentCmd.commands.Get(inputCmdStr)
			}
			if tmpCommand == nil {
				//jlog.Errorf("error: command u input not exist\n")
				return fmt.Errorf("command %s u input not exist", inputCmdStr)
			}
			//jlog.Error("currentCmd:", c.App.currentCmd)
			//jlog.Error(tmpCommand)
			//c.App.previousCmd = c.App.currentCmd // 记录切换前的command
			tmpCommand.previousCmd = c.App.currentCmd // 记录切换前的command
			c.App.currentCmd = tmpCommand
			// 获取命令
			//jlog.Warn(c.App.currentCommandStr)
			//jlog.Warn(tmpCommand.Name)
			//c.App.currentCommandStr = tmpCommand.Name
			// 设置prompt
			c.App.currentPrompt = c.App.config.Name + " " + tmpCommand.Name + "(" + tmpCommand.parentPath + ") >> "
			c.App.SetPrompt(c.App.currentPrompt)
			// 初始化jflagMaps
			if tmpCommand.jflagMaps == nil {
				tmpCommand.jflagMaps = make(FlagMap)
				// flag会使用默认值进行初始化，(arg只有list有默认空值，不能用这种方法)
				_, err := tmpCommand.flags.parse([]string{}, tmpCommand.jflagMaps)
				if err != nil {
					//jlog.Error(err)
					return err
				}
			}
			// JC 220512: 初始化jargMaps
			if tmpCommand.jargMaps == nil {
				tmpCommand.jargMaps = make(ArgMap)
			}
			// [+]220521: Add 设置自动填充参数
			//jlog.Error(c.App.currentCmd)
			if c.App.currentCmd == nil {
				//jlog.Error("app.commands.list:", c.App.commands.list)
				c.App.rl.Config.AutoComplete = newCompleter(&c.App.commands, c.App.currentCmd)
			} else {
				//jlog.Error("cmd.commands.list:",c.App.currentCmd.Name, c.App.currentCmd.commands.list)
				c.App.rl.Config.AutoComplete = newCompleter(&c.App.currentCmd.commands, c.App.currentCmd)
			}
			//for _, v := range c.App.currentCmd.commands.list {
			//	jlog.Warn("subCommand:", v.Name)
			//}
			return nil
		},
		isBuiltin: true,
	}
}

func core_show(a *App) *Command {
	return &Command{
		Name: "show",

		Aliases:   nil,
		Help:      "show options",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "show options",
		Flags:     nil,
		Args:      nil,
		Run: func(c *Context) error {
			// 获取当前command
			tmpCommand := c.App.currentCmd
			if tmpCommand == nil {
				//jlog.Errorf("error: command u input not exist\n")
				return fmt.Errorf("error: CurrentCommond is %v", tmpCommand)
			}
			// 输出当前flag
			t := table.NewWriter()
			t.AppendHeader(table.Row{"Name", "Value", "type", "Description"})
			//a.Printf("%-10v%-30v%-10v%-10v%v\n", "name", "value", "type", "isDefault", "description")
			//a.Println("=======================================================================")
			// JC 220512 遍历输出flag
			for _, v := range tmpCommand.flags.list {
				// JC 220514: 过滤掉help flag
				if v.Long == "help" {
					continue
				}
				if "slice" == reflect.TypeOf(tmpCommand.jflagMaps[v.Long].Value).Kind().String() {
					tmpStrSlice := make([]string, len(tmpCommand.jflagMaps[v.Long].Value.([]interface{})))
					for k2, v2 := range tmpCommand.jflagMaps[v.Long].Value.([]interface{}) {
						tmpStrSlice[k2] = fmt.Sprintf("%v", v2)
					}
					t.AppendRow(table.Row{v.Long, fmt.Sprintf("[%s]", strings.Join(tmpStrSlice, " ")), "flag", v.HelpArgs + ". " + v.Help})
					//a.Printf("%-10v%-30v%-10v%v\n", v.Long, "["+strings.Join(tmpStrSlice, " ")+"]", "flag", v.HelpArgs+". "+v.Help)
				} else {
					t.AppendRow(table.Row{v.Long, tmpCommand.jflagMaps[v.Long].Value, "flag", v.HelpArgs + ". " + v.Help})
					//a.Printf("%-10v%-30v%-10v%v\n", v.Long, tmpCommand.jflagMaps[v.Long].Value, "flag", v.HelpArgs+". "+v.Help)
				}
			}
			t.AppendSeparator()
			// JC 220512 遍历输出args
			for _, v := range tmpCommand.args.list {
				tmpStrSliceLen := 0
				tmpArg := reflect.Value{}
				tmpArgValue := ""
				// JC 220515: 判断是否设置
				if _, ok := tmpCommand.jargMaps[v.Name]; ok {
					// 判断是否list
					if v.isList {
						tmpArg = reflect.ValueOf(tmpCommand.jargMaps[v.Name].Value)
						tmpStrSliceLen = tmpArg.Len()
						tmpStrSlice := make([]string, tmpStrSliceLen)
						for k2, _ := range tmpStrSlice {
							tmpStrSlice[k2] = fmt.Sprintf("%v", tmpArg.Index(k2))
						}
						tmpArgValue = "[" + strings.Join(tmpStrSlice, " ") + "]"
					} else {
						tmpArgValue = fmt.Sprintf("%v", tmpCommand.jargMaps[v.Name].Value)
					}
					t.AppendRow(table.Row{v.Name, tmpArgValue, "arg", v.HelpArgs + ". " + v.Help})
					//a.Printf("%-10v%-30v%-10v%v\n", v.Name, tmpArgValue, "arg", v.HelpArgs+". "+v.Help)
				} else {
					t.AppendRow(table.Row{v.Name, "", "arg", v.HelpArgs + ". " + v.Help})
					//a.Printf("%-10v%-30v%-10v%v\n", v.Name, "", "arg", v.HelpArgs+". "+v.Help)
				}
			}
			a.Println(t.Render())
			return nil
		},
		isBuiltin: true,
		Completer: nil,
		parent:    nil,
		flags:     Flags{},
		args:      Args{},
		commands:  Commands{},
		jflagMaps: nil,
	}
}

func core_setf(a *App) *Command {
	return &Command{
		Name:      "setf",
		Aliases:   nil,
		Help:      "set flag",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "setf flag flagValue",
		//Flags:     nil,
		Args: func(a *Args) {
			a.String("argName", "setf flag flagValue")
			a.String("argValue", "setf flag flagValue")
		},
		Run: func(c *Context) error {
			// 获取当前command
			tmpCommand := c.App.currentCmd
			if tmpCommand == nil {
				return fmt.Errorf("error: CurrentCommond is %v", tmpCommand)
			}
			// 获取设置的参数
			argName := c.Args.String("argName")
			argValue := c.Args.String("argValue")
			//test1, err := shlex.Split(argValue2, true, false)
			//for _,v := range test1{
			//	jlog.Debug("_",v,"_")
			//}
			//if !strings.ContainsRune(arg, '=') {
			//	return fmt.Errorf("missing arg value")
			//}
			//argName := strings.Split(arg, "=")[0]
			//argValueTmp := strings.Split(arg, "=")[1:]
			//argValue := strings.Join(argValueTmp, "=")
			//argValueStr := strings.Join(argValue, "=")
			//jlog.Info("argValue:",trimQuotes(argValue))
			splitArgs, err := shlex.Split(argValue, true, false)
			if err != nil {
				return err
			}
			argValue = splitArgs[0]
			// 判断argName是否在当前命令的flag中
			for _, v := range tmpCommand.flags.list {
				if tmpCommand.jflagMaps == nil {
					tmpCommand.jflagMaps = make(FlagMap)
				}
				if argName == v.Long {
					// DONE 解析flag
					_, err := tmpCommand.flags.parse([]string{"--" + argName + "=" + argValue}, tmpCommand.jflagMaps)
					if err != nil {
						//jlog.Error(err)
						return err
					}
					break
				}
			}
			//jlog.Debug(tmpCommand.jflagMaps)
			return nil
		},
		isBuiltin: true,
		Completer: nil,
	}
}

func core_seta(a *App) *Command {
	return &Command{
		Name:      "seta",
		Aliases:   nil,
		Help:      "set arg",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "seta arg argValue",
		//Flags:     nil,
		Args: func(a *Args) {
			a.String("argName", "argName")
			a.String("argValue", "argValue")
		},
		Run: func(c *Context) error {
			// 获取当前command
			tmpCommand := c.App.currentCmd
			if tmpCommand == nil {
				return fmt.Errorf("error: CurrentCommond is %v", tmpCommand)
			}
			// 获取设置的参数
			argName := c.Args.String("argName")
			argValue := c.Args.String("argValue")
			//if !strings.ContainsRune(arg, '=') {
			//	return fmt.Errorf("missing arg value")
			//}
			//argName := strings.Split(arg, "=")[0]
			//argValueTmp := strings.Split(arg, "=")[1:]
			//argValue := strings.Join(argValueTmp, "=")
			//argValueStr := strings.Join(argValue,"=")
			//jlog.Info("argValue:", argValue)
			// 区分arg的类型
			var splitArgs = []string{argValue}
			// 枚举当前命令的arg
			if c.App.currentCmd != nil {
				for _, v := range c.App.currentCmd.args.list {
					if v.Name == argName {
						// 不是list类型
						//jlog.Error(argValue)
						if !v.isList {
							var err error
							splitArgs, err = shlex.Split(argValue, true, false)
							if err != nil {
								return err
							}
						}
					}
				}
			}
			argValue = splitArgs[0]
			//jlog.Info("argValue:", argValue)
			// 判断argName是否在当前命令的arg中
			for _, v := range tmpCommand.args.list {
				if argName == v.Name {
					if tmpCommand.jargMaps == nil {
						tmpCommand.jargMaps = make(ArgMap)
					}

					// 解析arg
					_, err := v.parser([]string{argValue}, tmpCommand.jargMaps)
					if err != nil {
						return err
					}
					//tmpCommand.jargMaps[v.Name] = &ArgMapItem{
					//	Value:     nil,
					//	IsDefault: false,
					//}
					break
				}

			}
			return nil
		},
		isBuiltin: true,
	}
}

func core_run(a *App) *Command {
	return &Command{
		Name: "run",

		Aliases:   nil,
		Help:      "run current command",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "run",
		Flags:     nil,
		Args:      nil,
		Run: func(c *Context) error {
			// 获取当前command
			tmpCommand := c.App.currentCmd
			if tmpCommand == nil {
				//jlog.Errorf("error: command u input not exist\n")
				return fmt.Errorf("error: CurrentCommond is %v,please use 'use <command>' first", tmpCommand)
			}
			// 判断是否设置了help=true
			if tmpCommand.jflagMaps["help"].Value == true {
				c.App.printCommandHelp(c.App, tmpCommand, c.App.isShell, true)
				return nil
			}
			// 执行前判断arg是否全部赋值
			for _, v := range tmpCommand.args.list {
				if _, ok := tmpCommand.jargMaps[v.Name]; !ok {
					return fmt.Errorf("请为所有arg类型的参数赋值")
				}
			}
			// 执行
			ctx := newContext(c.App, tmpCommand, tmpCommand.jflagMaps, tmpCommand.jargMaps)
			err := tmpCommand.Run(ctx)
			if err != nil {
				return err
			}
			return nil
		},
		isBuiltin: true,
		Completer: nil,
		parent:    nil,
		flags:     Flags{},
		args:      Args{},
		commands:  Commands{},
	}
}

func core_back(a *App) *Command {
	return &Command{
		Name: "back",

		Aliases:   nil,
		Help:      "back",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "back",
		Flags:     nil,
		Args: func(a *Args) {
			//a.String("args", "long flag name or all")
		},
		Run: func(c *Context) error {
			// JC 220521
			if c.App.currentCmd == nil {
				return nil
			}
			//jlog.Warn("currentCmd:", c.App.currentCmd)
			c.App.currentCmd = c.App.currentCmd.previousCmd
			if c.App.currentCmd == nil {
				c.App.currentPrompt = c.App.config.Name + " >> "
			} else {
				//jlog.Warn("previousCmd:", c.App.currentCmd)
				c.App.currentPrompt = c.App.config.Name + fmt.Sprintf(" %s(%s) >>", c.App.currentCmd.Name, c.App.currentCmd.parentPath)
			}
			//jlog.Warn("currentPrompt:", c.App.currentPrompt)
			c.App.SetPrompt(c.App.currentPrompt)
			// [+]220521: Add 设置自动填充参数
			if c.App.currentCmd == nil {
				//jlog.Error("app:", c.App.commands.list)
				c.App.rl.Config.AutoComplete = newCompleter(&c.App.commands, c.App.currentCmd)
			} else {
				// 初始化jflagMaps
				if c.App.currentCmd.jflagMaps == nil {
					c.App.currentCmd.jflagMaps = make(FlagMap)
					// flag会使用默认值进行初始化，(arg只有list有默认空值，不能用这种方法)
					_, err := c.App.currentCmd.flags.parse([]string{}, c.App.currentCmd.jflagMaps)
					if err != nil {
						//jlog.Error(err)
						return err
					}
				}
				// JC 220512: 初始化jargMaps
				if c.App.currentCmd.jargMaps == nil {
					c.App.currentCmd.jargMaps = make(ArgMap)
				}
				//jlog.Error(c.App.currentCmd.Name, c.App.currentCmd.commands.list)
				c.App.rl.Config.AutoComplete = newCompleter(&c.App.currentCmd.commands, c.App.currentCmd)
			}
			return nil
		},
		isBuiltin: true,
		Completer: nil,
		parent:    nil,
		flags:     Flags{},
		args:      Args{},
		commands:  Commands{},
		jflagMaps: nil,
	}
}

func core_unsetf(a *App) *Command {
	return &Command{
		Name: "unsetf",

		Aliases:   nil,
		Help:      "unset flag",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "unsetf <long flag name>|all",
		Flags:     nil,
		Args: func(a *Args) {
			a.String("args", "long flag name or all")
		},
		Run: func(c *Context) error {
			// 获取当前command
			tmpCommand := c.App.currentCmd
			if tmpCommand == nil {
				return fmt.Errorf("error: CurrentCommond is %v,please use 'use <command>' first \n", tmpCommand)
			}
			// 获取设置的参数
			arg := c.Args.String("args")
			// 初始化flag
			if arg == "all" { // 初始化每一个flag
				for _, v := range tmpCommand.flags.list {
					df := tmpCommand.flags.defaults[v.Long]
					df(tmpCommand.jflagMaps)
				}
				//jlog.Debug("unset all flag")
			} else { // 初始化指定flag
				for _, v := range tmpCommand.flags.list {
					if v.Long == arg {
						df := tmpCommand.flags.defaults[v.Long]
						df(tmpCommand.jflagMaps)
						return nil
					}
				}
			}
			return nil
		},
		isBuiltin: true,
		Completer: nil,
		parent:    nil,
		flags:     Flags{},
		args:      Args{},
		commands:  Commands{},
		jflagMaps: nil,
	}
}

func core_unseta(a *App) *Command {
	return &Command{
		Name: "unseta",

		Aliases:   nil,
		Help:      "unset arg ",
		LongHelp:  "",
		HelpGroup: jconfig.CORE_COMMAND_STR,
		Usage:     "unseta <arg name>|all",
		Flags:     nil,
		Args: func(a *Args) {
			a.String("args", "long arg name or all")
		},
		Run: func(c *Context) error {
			// 获取当前command
			tmpCommand := c.App.currentCmd
			if tmpCommand == nil {
				return fmt.Errorf("error: CurrentCommond is %v,please use 'use <command>' first \n", tmpCommand)
			}
			// 获取设置的参数
			arg := c.Args.String(("args"))
			// DONE 初始化arg
			if arg == "all" { // 初始化每一个arg
				for _, v := range tmpCommand.args.list {
					//删除对应arg的argMapItem
					delete(tmpCommand.jargMaps, v.Name)
					//df := tmpCommand.flags.defaults[v.Long]
					//df(tmpCommand.jflagMaps)
				}
				//jlog.Debug("unset all flag")
			} else { // 初始化指定flag
				for _, v := range tmpCommand.args.list {
					if v.Name == arg {
						//df := tmpCommand.flags.defaults[v.Long]
						//df(tmpCommand.jflagMaps)
						delete(tmpCommand.jargMaps, v.Name)
						return nil
					}
				}
			}
			return nil
		},
		isBuiltin: true,
		Completer: nil,
		parent:    nil,
		flags:     Flags{},
		args:      Args{},
		commands:  Commands{},
		jflagMaps: nil,
	}
}
