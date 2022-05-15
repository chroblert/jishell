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
	commands       *Commands
	currentCommand string
}

func newCompleter(commands *Commands, s string) *completer {
	//jlog.Error(commands,s)
	return &completer{
		commands:       commands,
		currentCommand: s,
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
		words = strings.Fields(string(line)) // fallback
	}
	//jlog.Error("words:",words)
	prefix := ""
	// 如果字符串列表不为空，且pos大于1，并且最后一个字符不为空格
	// prefix为空白分隔的words的最后一个字符串
	// words为其余的字符串列表
	// 若最后一个字符为空格，则prefix为空字符串；words为字符串列表
	if len(words) > 0 && pos >= 1 && line[pos-1] != ' ' {
		prefix = words[len(words)-1]
		words = words[:len(words)-1]
	}
	//jlog.Error(len(line),"line2:",line,", pos:",pos)
	//jlog.Error(len(prefix),"prefix:",prefix)
	//jlog.Error(len(words),"words2:",words)

	// Simple hack to allow auto completion for help.
	if len(words) > 0 && words[0] == "help" {
		words = words[1:]
	}

	var (
		cmds  *Commands
		flags *Flags
		//args 		*Args
		suggestions [][]rune
	)

	// Find the last commands list.
	if len(words) == 0 {
		cmds = c.commands
	} else { // 子命令 xxx形式
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

		cmds = &cmd.commands
		flags = &cmd.flags
		//args = &cmd.args
	}

	// JC 220515:自动补全要切换的命令
	if len(words) > 0 && words[0] == "use" {
		for _, v := range c.commands.list {
			if v.CMDPath != "" {
				if len(prefix) > 0 {
					if strings.HasPrefix(v.CMDPath+"/"+v.Name, prefix) {
						suggestions = append(suggestions, []rune(strings.TrimPrefix(v.CMDPath+"/"+v.Name, prefix))) // JC 220515
					}
				} else {
					suggestions = append(suggestions, []rune(v.CMDPath+"/"+v.Name)) // JC 220515
				}
			}
		}
	}
	// [+]210530: Add setf 自动补全当前cmd的flag
	// [·]220515: update
	if len(words) > 0 && words[0] == "setf" {
		// 获取当前的命令
		for _, v := range c.commands.list {
			if v.Name == c.currentCommand {
				for _, v2 := range v.flags.list {
					if len(prefix) > 0 {
						if strings.HasPrefix(v2.Long, prefix) {
							suggestions = append(suggestions, []rune(strings.TrimPrefix(v2.Long, prefix)))
						}
					} else {
						suggestions = append(suggestions, []rune(v2.Long))
					}
				}
				break
			}
		}
	}
	// end
	// [+]220515: Add seta 自动补全当前cmd的arg
	if len(words) > 0 && words[0] == "seta" {
		// 获取当前的命令
		for _, v := range c.commands.list {
			if v.Name == c.currentCommand {
				for _, v2 := range v.args.list {
					if len(prefix) > 0 {
						if strings.HasPrefix(v2.Name, prefix) {
							suggestions = append(suggestions, []rune(strings.TrimPrefix(v2.Name, prefix)))
						}
					} else {
						suggestions = append(suggestions, []rune(v2.Name))
					}
				}
				break
			}
		}
	}
	// end
	// [+]220515: Add unsetf 自动补全当前cmd的flag
	if len(words) > 0 && words[0] == "unsetf" {
		// 获取当前的命令
		for _, v := range c.commands.list {
			if v.Name == c.currentCommand {
				for _, v2 := range v.flags.list {
					if len(prefix) > 0 {
						if strings.HasPrefix(v2.Long, prefix) {
							suggestions = append(suggestions, []rune(strings.TrimPrefix(v2.Long, prefix)))
						}
					} else {
						suggestions = append(suggestions, []rune(v2.Long))
					}
				}
				break
			}
		}
	}
	// end
	// [+]220515: Add unseta 自动补全当前cmd的arg
	if len(words) > 0 && words[0] == "unseta" {
		// 获取当前的命令
		for _, v := range c.commands.list {
			if v.Name == c.currentCommand {
				for _, v2 := range v.args.list {
					if len(prefix) > 0 {
						if strings.HasPrefix(v2.Name, prefix) {
							suggestions = append(suggestions, []rune(strings.TrimPrefix(v2.Name, prefix)))
						}
					} else {
						suggestions = append(suggestions, []rune(v2.Name))
					}
				}
				break
			}
		}
	}
	// end
	if len(prefix) > 0 {
		// 看有没有子命令
		for _, cmd := range cmds.list {
			if strings.HasPrefix(cmd.Name, prefix) {
				suggestions = append(suggestions, []rune(strings.TrimPrefix(cmd.Name, prefix)))
			}
			for _, a := range cmd.Aliases {
				if strings.HasPrefix(a, prefix) {
					suggestions = append(suggestions, []rune(strings.TrimPrefix(a, prefix)))
				}
			}
		}

		// 自动补全flag，默认显示long flag
		if flags != nil {
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
	} else {

		for _, cmd := range cmds.list {
			suggestions = append(suggestions, []rune(cmd.Name))
		}
		if flags != nil {
			for _, f := range flags.list {
				if f.Long != "" {
					suggestions = append(suggestions, []rune("--"+f.Long))
				} else if len(f.Short) > 0 {
					suggestions = append(suggestions, []rune("-"+f.Short))
				}
			}
		}
	}
	return suggestions, len(prefix)
}
