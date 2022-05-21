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
	"strings"

	shlex "github.com/chroblert/go-shlex"
)

type completer struct {
	commands   *Commands
	currentCmd *Command
}

func newCompleter(commands *Commands, currentCmd *Command) *completer {
	//jlog.Error("new completer")
	return &completer{
		commands:   commands,
		currentCmd: currentCmd,
	}
}

func (c *completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// Discard anything after the cursor position.
	// This is similar behaviour to shell/bash.
	//jlog.Error(string(line),pos)
	line = line[:pos]

	var words []string
	// 以空白字符进行分隔，若无报错，则words值为空白字符分隔的字符串列表
	if w, err := shlex.Split(string(line), true, false); err == nil {
		words = w
	} else {
		//jlog.Error("error:", err)
		words = strings.Fields(string(line)) // fallback
	}
	//jlog.Error("words:",words)
	prefix := ""
	// 如果字符串列表不为空，且pos大于1，并且最后一个字符不为空格
	// prefix为空白分隔的words的最后一个字符串
	// words为其余的字符串列表
	// 若最后一个字符为空格，则prefix为空字符串；words为字符串列表
	//jlog.Warn("-:", len(words), words, ",line:", fmt.Sprintf("_%s_", string(line)))
	if len(words) > 0 && pos >= 1 && line[pos-1] != ' ' {
		prefix = words[len(words)-1]
		words = words[:len(words)-1]
	}
	//jlog.Warn("--:", len(words), words, ",--prefix:", prefix)
	//jlog.Error(len(line),"line2:",line,", pos:",pos)
	//jlog.Error(len(prefix),"prefix:",prefix)
	//jlog.Error(len(words),"words2:",words)
	var (
		cmds              []*Command
		flags             *Flags
		args              *Args
		suggestions       [][]rune
		firstIsBuiltInCmd bool
	)
	firstIsBuiltInCmd = true
	if len(words) == 0 {
		for _, v := range c.commands.list {
			//jlog.Info("cmd:", v.Name, v.isBuiltin)
			if !v.isBuiltin {
				cmds = append(cmds, v)
			}
		}
	} else {
		switch words[0] {
		case "help":
			for _, v := range c.commands.list {
				//jlog.Info("cmd:", v.Name, v.isBuiltin)
				if !v.isBuiltin {
					cmds = append(cmds, v)
				}
			}
		case "use":
			for _, v := range c.commands.list {
				//jlog.Info("cmd:", v.Name, v.isBuiltin)
				if !v.isBuiltin {
					cmds = append(cmds, v)
				}
			}
		case "seta":
			if c.currentCmd != nil {
				args = &c.currentCmd.args
			}
		case "setf":
			if c.currentCmd != nil {
				flags = &c.currentCmd.flags
			}
		case "unseta":
			if c.currentCmd != nil {
				args = &c.currentCmd.args
			}
		case "unsetf":
			if c.currentCmd != nil {
				flags = &c.currentCmd.flags
			}
		default: // 非内置命令
			firstIsBuiltInCmd = false
			// 子命令 xxx形式
			// 目的在于找到最后一个命令
			cmd, rest, err := c.commands.FindCommand(words)
			if err != nil || cmd == nil {
				return
			}
			// Call the custom completer if present.
			if cmd.Completer != nil {
				words = cmd.Completer(prefix, rest)
				for _, w := range words {
					suggestions = append(suggestions, []rune(strings.TrimPrefix(w, prefix)))
				}
				return suggestions, len(prefix)
			}
			// No rest must be there.
			if len(rest) != 0 {
				return
			}
			cmds = cmd.commands.list
			flags = &cmd.flags
			args = &cmd.args
			//jlog.Warn("cmds1:", len(cmds), cmds)
		}
	}

	//jlog.Warn("words1:", len(words), words)
	//for _, v := range c.commands.list {
	//	jlog.Warn("<in use> subCommand:", v.Name)
	//}
	//if args != nil {
	//	for _, v := range args.list {
	//		jlog.Warn("<arg> :", v.Name)
	//	}
	//}
	//if flags != nil {
	//	for _, v := range flags.list {
	//		jlog.Warn("<flag> :", v.Long)
	//	}
	//}

	if len(prefix) > 0 {
		// 看有没有子命令
		if cmds != nil {
			for _, cmd := range cmds {
				if strings.HasPrefix(cmd.Name, prefix) {
					suggestions = append(suggestions, []rune(strings.TrimPrefix(cmd.Name, prefix)))
				} else {
					// 防止重复输出同一个命令的名称和别名
					for _, a := range cmd.Aliases {
						if strings.HasPrefix(a, prefix) {
							suggestions = append(suggestions, []rune(strings.TrimPrefix(a, prefix)))
						}
					}
				}

			}
		}

		// 自动补全flag，默认显示long flag
		if flags != nil {
			// 第一个字符串是内置命令，则只使用长模式
			if firstIsBuiltInCmd {
				for _, f := range flags.list {
					long := f.Long
					if len(prefix) < len(long) && strings.HasPrefix(long, prefix) {
						suggestions = append(suggestions, []rune(strings.TrimPrefix(long, prefix)))
					}
				}
			} else {
				for _, f := range flags.list {
					long := "--" + f.Long
					if len(prefix) < len(long) && strings.HasPrefix(long, prefix) {
						suggestions = append(suggestions, []rune(strings.TrimPrefix(long, prefix)))
						continue
					}
					if len(f.Short) > 0 {
						short := "-" + f.Short
						if len(prefix) < len(short) && strings.HasPrefix(short, prefix) {
							suggestions = append(suggestions, []rune(strings.TrimPrefix(short, prefix)))
						}
					}
				}
			}

		}
		// 自动补全arg，默认显示long flag
		if firstIsBuiltInCmd && args != nil {
			for _, a := range args.list {
				long := a.Name
				if len(prefix) < len(long) && strings.HasPrefix(long, prefix) {
					suggestions = append(suggestions, []rune(strings.TrimPrefix(long, prefix)))
					continue
				}
			}
		}
	} else {
		if cmds != nil {
			for _, cmd := range cmds {
				suggestions = append(suggestions, []rune(cmd.Name))
			}
		}

		if flags != nil {
			if firstIsBuiltInCmd {
				for _, f := range flags.list {
					if f.Long != "" {
						suggestions = append(suggestions, []rune(f.Long))
					}
				}
			} else {
				for _, f := range flags.list {
					if f.Long != "" {
						suggestions = append(suggestions, []rune("--"+f.Long))
					} else if len(f.Short) > 0 {
						suggestions = append(suggestions, []rune("-"+f.Short))
					}
				}
			}
		}
		// 自动补全arg，默认显示long flag
		if firstIsBuiltInCmd && args != nil {
			for _, a := range args.list {
				long := a.Name
				if len(prefix) < len(long) && strings.HasPrefix(long, prefix) {
					suggestions = append(suggestions, []rune(strings.TrimPrefix(long, prefix)))
					continue
				}
			}
		}
	}
	return suggestions, len(prefix)
}
