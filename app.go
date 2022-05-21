/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer [roland.singer@deserbit.com]
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package jishell

import (
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"io"
	"os"
	"reflect"
	"strings"

	shlex "github.com/chroblert/go-shlex"
	"github.com/desertbit/closer/v3"
	"github.com/desertbit/readline"
	"github.com/fatih/color"
)

// App is the entrypoint.
type App struct {
	closer.Closer

	rl            *readline.Instance
	config        *Config
	commands      Commands
	isShell       bool
	currentPrompt string

	flags   Flags
	flagMap FlagMap

	args Args

	initHook  func(a *App, flags FlagMap) error
	shellHook func(a *App) error

	printHelp        func(a *App, shell bool)
	printCommandHelp func(a *App, cmd *Command, shell bool)
	interruptHandler func(a *App, count int)
	printASCIILogo   func(a *App)

	// JC0o0l Add
	//currentCommandStr string
	currentCmd  *Command // JC 220520 存放当前的Command，初始为nil
	previousCmd *Command // JC 220520 存放use切换前的Command，初始为nil
}

// New creates a new app.
// Panics if the config is invalid.
func New(c *Config) (a *App) {
	// Prepare the config.
	c.SetDefaults()
	err := c.Validate()
	if err != nil {
		panic(err)
	}

	// APP.
	a = &App{
		Closer:           closer.New(),
		config:           c,
		currentPrompt:    c.prompt(),
		flagMap:          make(FlagMap),
		printHelp:        defaultPrintHelp,
		printCommandHelp: defaultPrintCommandHelp,
		interruptHandler: defaultInterruptHandler,
		//currentCommandStr: c.CurrentCmdStr,
		currentCmd: nil,
	}

	// Register the builtin flags.
	a.flags.Bool("h", "help", false, "display help")
	a.flags.BoolL("nocolor", false, "disable color output")
	//a.flags.Bool("i", "interactive", true, "enable interactive mode")

	// Register the user flags, if present.
	if c.Flags != nil {
		c.Flags(&a.flags)
	}

	return
}

// SetPrompt sets a new prompt.
func (a *App) SetPrompt(p string) {
	if !a.config.NoColor {
		p = a.config.PromptColor.Sprint(p)
	}
	a.currentPrompt = p
}

// GetPromp get a prompt string
func (a *App) GetPrompt() string {
	return a.currentPrompt
}

//
//func (a *App) SetCurrentCommand(c string) {
//	a.currentCommandStr = c
//}
//
//func (a *App) GetCurrentCommand() string {
//	return a.currentCommandStr
//}

// SetDefaultPrompt resets the current prompt to the default prompt as
// configured in the config.
func (a *App) SetDefaultPrompt() {
	a.currentPrompt = a.config.prompt()
}

// IsShell indicates, if this is a shell session.
func (a *App) IsShell() bool {
	return a.isShell
}

// Config returns the app's config value.
func (a *App) Config() *Config {
	return a.config
}

// Commands returns the app's commands.
// Access is not thread-safe. Only access during command execution.
func (a *App) Commands() *Commands {
	return &a.commands
}

// PrintError prints the given error.
func (a *App) PrintError(err error) {
	if a.config.NoColor {
		a.Printf("error: %v\n", err)
	} else {
		a.config.ErrorColor.Fprint(a, "error: ")
		a.Printf("%v\n", err)
	}
}

// Print writes to terminal output.
// Print writes to standard output if terminal output is not yet active.
func (a *App) Print(args ...interface{}) (int, error) {
	return fmt.Fprint(a, args...)
}

// Printf formats according to a format specifier and writes to terminal output.
// Printf writes to standard output if terminal output is not yet active.
func (a *App) Printf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(a, format, args...)
}

// Println writes to terminal output followed by a newline.
// Println writes to standard output if terminal output is not yet active.
func (a *App) Println(args ...interface{}) (int, error) {
	return fmt.Fprintln(a, args...)
}

// OnInit sets the function which will be executed before the first command
// is executed. App flags can be handled here.
func (a *App) OnInit(f func(a *App, flags FlagMap) error) {
	a.initHook = f
}

// OnShell sets the function which will be executed before the shell starts.
func (a *App) OnShell(f func(a *App) error) {
	a.shellHook = f
}

// SetInterruptHandler sets the interrupt handler function.
func (a *App) SetInterruptHandler(f func(a *App, count int)) {
	a.interruptHandler = f
}

// SetPrintHelp sets the print help function.
func (a *App) SetPrintHelp(f func(a *App, shell bool)) {
	a.printHelp = f
}

// SetPrintCommandHelp sets the print help function for a single command.
func (a *App) SetPrintCommandHelp(f func(a *App, c *Command, shell bool)) {
	a.printCommandHelp = f
}

// SetPrintASCIILogo sets the function to print the ASCII logo.
func (a *App) SetPrintASCIILogo(f func(a *App)) {
	a.printASCIILogo = func(a *App) {
		if !a.config.NoColor {
			a.config.ASCIILogoColor.Set()
			defer color.Unset()
		}
		f(a)
	}
}

// Write to the underlying output, using readline if available.
func (a *App) Write(p []byte) (int, error) {
	return a.Stdout().Write(p)
}

// Stdout returns a writer to Stdout, using readline if available.
// Note that calling before Run() will return a different instance.
func (a *App) Stdout() io.Writer {
	if a.rl != nil {
		return a.rl.Stdout()
	}
	return os.Stdout
}

// Stderr returns a writer to Stderr, using readline if available.
// Note that calling before Run() will return a different instance.
func (a *App) Stderr() io.Writer {
	if a.rl != nil {
		return a.rl.Stderr()
	}
	return os.Stderr
}

// AddCommand adds a new command.
// Panics on error.
func (a *App) AddCommand(cmd *Command) {
	a.addCommand(cmd, true)
}

// addCommand adds a new command.
// If addHelpFlag is true, a help flag is automatically
// added to the command which displays its usage on use.
// Panics on error.
func (a *App) addCommand(cmd *Command, addHelpFlag bool) {
	err := cmd.validate()
	if err != nil {
		panic(err)
	}
	cmd.parentPath = "/" // JC 220521 为一级子命令设置prompt
	cmd.registerFlagsAndArgs(addHelpFlag)

	a.commands.Add(cmd)
}

// RunCommand runs a single command.
func (a *App) RunCommand(args []string) error {
	// Parse the arguments string and obtain the command path to the root,
	// and the command flags.
	var (
		cmds []*Command
		fg   FlagMap
		//args []string
		err error
	)
	if a.currentCmd == nil {
		cmds, fg, args, err = a.commands.parse(args, a.flagMap, false)
	} else {
		var tmpCommands = Commands{}
		for _, v := range a.commands.list {
			tmpCommands.Add(v)
		}
		for _, v := range a.currentCmd.commands.list {
			tmpCommands.Add(v)
		}
		//tmpCommands = a.commands.
		cmds, fg, args, err = tmpCommands.parse(args, a.flagMap, false)
	}
	if err != nil {
		return err
	} else if len(cmds) == 0 {
		return fmt.Errorf("unknown command, try 'help'")
	}

	// The last command is the final command.
	cmd := cmds[len(cmds)-1]

	// Print the command help if the command run function is nil or if the help flag is set.
	if fg.Bool("help") || cmd.Run == nil {
		a.printCommandHelp(a, cmd, a.isShell)
		return nil
	}

	// Parse the arguments.
	cmdArgMap := make(ArgMap)
	args, err = cmd.args.parse(args, cmdArgMap)
	if err != nil {
		return err
	}
	// Check, if values from the argument string are not consumed (and therefore invalid).
	if len(args) > 0 {
		return fmt.Errorf("invalid usage of command '%s' (unconsumed input '%s'), try 'help'", cmd.Name, strings.Join(args, " "))
	}

	// Create the context and pass the rest args.
	ctx := newContext(a, cmd, fg, cmdArgMap)

	// Run the command.
	err = cmd.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Run the application and parse the command line arguments.
// This method blocks.
func (a *App) Run() (err error) {
	defer a.Close()

	// Sort all commands by their name.
	a.commands.SortRecursive()

	// TODO 有没有一种方法，能够接收控制台上输入的所有字符，而不是去掉双引号后的值
	// Remove the program name from the args.
	args := os.Args
	if len(args) > 0 {
		args = args[1:]
	}
	// 如果可执行程序后跟的第一个参数是 -i,则进入交互模式；默认是控制台模式
	//if len(args) > 0 && args[0] == "-i"{
	//	a.isShell = true
	//	args = args[1:]
	//}
	// Parse the app command line flags.
	args, err = a.flags.parse(args, a.flagMap)
	if err != nil {
		return err
	}

	// Check if nocolor was set.
	a.config.NoColor = a.flagMap.Bool("nocolor")
	// Determine if this is a shell session.
	a.isShell = len(args) == 0
	// JC 220520 获取-i flag值
	//a.isShell = a.flagMap.Bool("interactive")
	// JC 220520 再根据是否有别的参数，来设置是否为shell模式
	//if len(args) > 0 {
	//	a.isShell = false
	//}
	//jlog.Error(len(args),args)
	//jlog.Error(a.config.NoColor,a.isShell)

	// Add general builtin commands.
	a.addCommand(&Command{
		Name: "help",

		Help:      "use 'help [command]' for command help",
		HelpGroup: "Core Commands",
		Args: func(a *Args) {
			a.StringList("command", "the name of the command")
		},
		Run: func(c *Context) error {
			args := c.Args.StringList("command")
			if len(args) == 0 {
				a.printHelp(a, a.isShell)
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
			a.printCommandHelp(a, cmd, a.isShell)
			return nil
		},
		isBuiltin: true,
	}, false)

	// Check if help should be displayed.
	if a.flagMap.Bool("help") {
		a.printHelp(a, false)
		return nil
	}

	// Add shell builtin commands.
	// Ensure to add all commands before running the init hook.
	// If the init hook does something with the app commands, then these should also be included.
	if a.isShell {
		// Add shell builtin commands.
		a.AddCommand(&Command{
			Name: "exit",

			Help:      "exit the shell",
			HelpGroup: "Core Commands",
			Run: func(c *Context) error {
				c.Stop()
				return nil
			},
			isBuiltin: true,
		})
		a.AddCommand(&Command{
			Name: "clear",

			Help:      "clear the screen",
			HelpGroup: "Core Commands",
			Run: func(c *Context) error {
				readline.ClearScreen(a.rl)
				return nil
			},
			isBuiltin: true,
		})
		// 添加use命令
		a.AddCommand(&Command{
			Name: "use",

			Aliases:   nil,
			Help:      "switch command",
			LongHelp:  "",
			HelpGroup: "Core Commands",
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
		})
		// 添加show命令
		a.AddCommand(&Command{
			Name: "show",

			Aliases:   nil,
			Help:      "show options",
			LongHelp:  "",
			HelpGroup: "Core Commands",
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
				a.Printf("%-10v%-30v%-10v%-10v%v\n", "name", "value", "type", "isDefault", "description")
				a.Println("=======================================================================")
				// JC 220512 遍历输出flag
				for _, v := range tmpCommand.flags.list {
					// JC 220514: 过滤掉help flag
					if v.Long == "help" {
						continue
					}
					// TODO 待完善: 是否能够自适应
					if "slice" == reflect.TypeOf(tmpCommand.jflagMaps[v.Long].Value).Kind().String() {
						tmpStrSlice := make([]string, len(tmpCommand.jflagMaps[v.Long].Value.([]interface{})))
						for k2, v2 := range tmpCommand.jflagMaps[v.Long].Value.([]interface{}) {
							tmpStrSlice[k2] = fmt.Sprintf("%v", v2)
						}
						a.Printf("%-10v%-30v%-10v%v\n", v.Long, "["+strings.Join(tmpStrSlice, " ")+"]", "flag", v.HelpArgs+". "+v.Help)
					} else {
						a.Printf("%-10v%-30v%-10v%v\n", v.Long, tmpCommand.jflagMaps[v.Long].Value, "flag", v.HelpArgs+". "+v.Help)
					}
				}
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
						a.Printf("%-10v%-30v%-10v%v\n", v.Name, tmpArgValue, "arg", v.HelpArgs+". "+v.Help)
					} else {
						a.Printf("%-10v%-30v%-10v%v\n", v.Name, "", "arg", v.HelpArgs+". "+v.Help)
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
		})
		// 添加setf命令
		a.AddCommand(&Command{
			Name:      "setf",
			Aliases:   nil,
			Help:      "set flag",
			LongHelp:  "",
			HelpGroup: "Core Commands",
			Usage:     "setf flag=flagValue",
			//Flags:     nil,
			Args: func(a *Args) {
				a.String("args", "flag=flagValue")
			},
			Run: func(c *Context) error {
				// 获取当前command
				tmpCommand := c.App.currentCmd
				if tmpCommand == nil {
					return fmt.Errorf("error: CurrentCommond is %v", tmpCommand)
				}
				// 获取设置的参数
				arg := c.Args.String("args")
				if !strings.ContainsRune(arg, '=') {
					return fmt.Errorf("missing arg value")
				}
				argName := strings.Split(arg, "=")[0]
				argValue := strings.Split(arg, "=")[1:]
				argValueStr := strings.Join(argValue, "=")
				// 判断argName是否在当前命令的flag中
				for _, v := range tmpCommand.flags.list {
					if tmpCommand.jflagMaps == nil {
						tmpCommand.jflagMaps = make(FlagMap)
					}
					if argName == v.Long {
						// DONE 解析flag
						_, err := tmpCommand.flags.parse([]string{"--" + argName + "=" + argValueStr}, tmpCommand.jflagMaps)
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
		})
		// 添加setf命令
		a.AddCommand(&Command{
			Name:      "seta",
			Aliases:   nil,
			Help:      "set arg",
			LongHelp:  "",
			HelpGroup: "Core Commands",
			Usage:     "seta arg=argValue",
			//Flags:     nil,
			Args: func(a *Args) {
				a.String("args", "arg=argValue")
			},
			Run: func(c *Context) error {
				// 获取当前command
				tmpCommand := c.App.currentCmd
				if tmpCommand == nil {
					return fmt.Errorf("error: CurrentCommond is %v", tmpCommand)
				}
				// 获取设置的参数
				arg := c.Args.String("args")
				if !strings.ContainsRune(arg, '=') {
					return fmt.Errorf("missing arg value")
				}
				argName := strings.Split(arg, "=")[0]
				argValue := strings.Split(arg, "=")[1]
				//argValueStr := strings.Join(argValue,"=")
				jlog.Info("argValue:", argValue)
				// 区分arg的类型
				var splitArgs = []string{argValue}
				// 枚举当前命令的arg
				if c.App.currentCmd != nil {
					for _, v := range c.App.currentCmd.args.list {
						if v.Name == argName {
							// 不是list类型
							jlog.Error(v.isList)
							if !v.isList {
								splitArgs, err = shlex.Split(argValue, true, false)
								if err != nil {
									return err
								}
							}
						}
					}
				}
				argValue = splitArgs[0]
				jlog.Info("argValue:", argValue)
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
		})
		// 添加run命令
		a.AddCommand(&Command{
			Name: "run",

			Aliases:   nil,
			Help:      "run current command",
			LongHelp:  "",
			HelpGroup: "Core Commands",
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
					c.App.printCommandHelp(c.App, tmpCommand, c.App.isShell)
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
				err = tmpCommand.Run(ctx)
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
		})
		// 添加run命令
		a.AddCommand(&Command{
			Name: "run",

			Aliases:   nil,
			Help:      "run current command",
			LongHelp:  "",
			HelpGroup: "Core Commands",
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
					c.App.printCommandHelp(c.App, tmpCommand, c.App.isShell)
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
				err = tmpCommand.Run(ctx)
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
		})
		// 添加back命令
		a.AddCommand(&Command{
			Name: "back",

			Aliases:   nil,
			Help:      "back",
			LongHelp:  "",
			HelpGroup: "Core Commands",
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
		})
		// 添加unsetf命令
		a.AddCommand(&Command{
			Name: "unsetf",

			Aliases:   nil,
			Help:      "unset flag",
			LongHelp:  "",
			HelpGroup: "Core Commands",
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
		})
		// 添加unseta命令
		a.AddCommand(&Command{
			Name: "unseta",

			Aliases:   nil,
			Help:      "unset arg ",
			LongHelp:  "",
			HelpGroup: "Core Commands",
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
		})

	}
	// Run the init hook.
	if a.initHook != nil {
		err = a.initHook(a, a.flagMap)
		if err != nil {
			return err
		}
	}

	// Check if a command chould be executed in non-interactive mode.
	if !a.isShell {
		return a.RunCommand(args)
	}

	// Create the readline instance.
	a.rl, err = readline.NewEx(&readline.Config{
		Prompt:                 a.currentPrompt,
		HistorySearchFold:      true, // enable case-insensitive history searching
		DisableAutoSaveHistory: true,
		HistoryFile:            a.config.HistoryFile,
		HistoryLimit:           a.config.HistoryLimit,
		AutoComplete:           newCompleter(&a.commands, nil),
		VimMode:                a.config.VimMode,
	})

	if err != nil {
		return err
	}
	a.OnClose(a.rl.Close)

	// Run the shell hook.
	if a.shellHook != nil {
		err = a.shellHook(a)
		if err != nil {
			return err
		}
	}

	// Print the ASCII logo.
	if a.printASCIILogo != nil {
		a.printASCIILogo(a)
	}

	// Run the shell.
	return a.runShell()
}

func (a *App) runShell() error {
	var interruptCount int
	var lines []string
	multiActive := false

Loop:
	for !a.IsClosing() {
		// Set the prompt.
		if multiActive {
			a.rl.SetPrompt(a.config.multiPrompt())
		} else {
			a.rl.SetPrompt(a.currentPrompt)
		}
		multiActive = false

		// Readline.
		line, err := a.rl.Readline()
		//jlog.Error("line:",line)
		if err != nil {
			if err == readline.ErrInterrupt {
				interruptCount++
				a.interruptHandler(a, interruptCount)
				continue Loop
			} else if err == io.EOF {
				return nil
			} else {
				return err
			}
		}

		// Reset the interrupt count.
		interruptCount = 0

		// Handle multiline input.
		if strings.HasSuffix(line, "\\") {
			multiActive = true
			line = strings.TrimSpace(line[:len(line)-1]) // Add without suffix and trim spaces.
			lines = append(lines, line)
			continue Loop
		}
		lines = append(lines, strings.TrimSpace(line))

		line = strings.Join(lines, "")
		line = strings.TrimSpace(line)
		lines = lines[:0]

		// Skip if the line is empty.
		if len(line) == 0 {
			continue Loop
		}

		// Save command history.
		err = a.rl.SaveHistory(line)
		if err != nil {
			a.PrintError(err)
			continue Loop
		}

		// Split the line to args.
		args, err := shlex.Split(line, true, true, ' ')
		//jlog.Error("args:",len(args),args)
		if err != nil {
			a.PrintError(fmt.Errorf("invalid args: %v", err))
			continue Loop
		}
		// Execute the command.
		err = a.RunCommand(args)
		if err != nil {
			a.PrintError(err)
			// Do not continue the Loop here. We want to handle command changes below.
			//continue Loop
		}
		// Sort the commands again if they have changed (Add or remove action).
		if a.commands.hasChanged() {
			a.commands.SortRecursive()
			a.commands.unsetChanged()
		}
	}

	return nil
}
