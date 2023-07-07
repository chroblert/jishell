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
	"github.com/chroblert/go-shlex"
	"strconv"
	"time"
)

// The parseArgFunc describes a func that parses from the given command line arguments
// the values for its argument and saves them to the ArgMap.
// It returns the not-consumed arguments and an error.
type parseArgFunc func(args []string, res ArgMap) ([]string, error)

type argItem struct {
	Name     string
	Help     string
	HelpArgs string
	Default  interface{}

	parser   parseArgFunc
	isList   bool
	optional bool
	listMin  int
	listMax  int
}

// Args holds all the registered args.
type Args struct {
	list []*argItem
}

func (a *Args) register(
	name, help, helpArgs string,
	isList bool,
	pf parseArgFunc,
	opts ...ArgOption,
) {
	// Validate.
	if name == "" {
		panic("empty argument name")
	}
	// JC 220513: 允许没有help信息
	//else if help == "" {
	//	panic(fmt.Errorf("missing help message for argument '%s'", name))
	//}

	// Ensure the name is unique.
	for _, ai := range a.list {
		if ai.Name == name {
			panic(fmt.Errorf("argument '%s' registered twice", name))
		}
	}

	// Create the item.
	item := &argItem{
		Name:     name,
		Help:     help,
		HelpArgs: helpArgs,
		parser:   pf,
		isList:   isList,
		optional: isList,
		listMin:  -1,
		listMax:  -1,
	}

	// Apply options.
	// Afterwards, we can make some final checks.
	for _, opt := range opts {
		opt(item)
	}

	if item.isList && item.listMax > 0 && item.listMax < item.listMin {
		panic("max must not be less than min for list arguments")
	}

	// JC 220514: 取消只允许一个list的限制，取消list在最后一个参数的限制
	//if !a.empty() {
	//	last := a.list[len(a.list)-1]
	//
	//	// Check, if a list argument has been supplied already.
	//	if last.isList {
	//		panic("list argument has been registered, nothing can come after it")
	//	}
	//
	//	// Check, that after an optional argument no mandatory one follows.
	//	if !item.optional && last.optional {
	//		panic("mandatory argument not allowed after optional one")
	//	}
	//}

	a.list = append(a.list, item)
}

// empty returns true, if the args are empty.
func (a *Args) empty() bool {
	return len(a.list) == 0
}

// 如果运行正常，则len([]string{})应该为0
func (a *Args) parse(args []string, res ArgMap) ([]string, error) {
	// Iterate over all arguments that have been registered.
	// There must be either a default value or a value available,
	// otherwise the argument is missing.
	var err error
	// JC 220514: 剩余的参数个数，应该与注册的arg的个数相等 len(a.list)
	//if len(args) < len(a.list){
	//	jlog.Fatal(len(args),len(a.list),"len(args)<len(a.list)")
	//	return args,fmt.Errorf("传入的参数个数小于设定的参数个数")
	//}
	for _, item := range a.list {

		// If no arguments are left, simply set the default values.
		if len(args) == 0 {
			// Check, if the argument is mandatory.
			if !item.optional {
				return nil, fmt.Errorf("missing argument '%s'", item.Name)
			}

			// Register its default value.
			res[item.Name] = &ArgMapItem{Value: item.Default, IsDefault: true}
			continue
		}
		// JC 220514: 调换一下顺序
		// If it is a list argument, it will consume the rest of the input.
		// Check that it matches its range.
		// 用于校验
		if item.isList {
			// List以,分隔
			args0List, err := shlex.Split(args[0], true, false, ',')
			if err != nil {
				return nil, err
			}
			if len(args0List) < item.listMin {
				return nil, fmt.Errorf("argument '%s' requires at least %d element(s)", item.Name, item.listMin)
			}
			if item.listMax > 0 && len(args0List) > item.listMax {
				return nil, fmt.Errorf("argument '%s' requires at most %d element(s)", item.Name, item.listMax)
			}
		}

		args, err = item.parser(args, res)
		if err != nil {
			return nil, err
		}
	}

	return args, nil
}

// String registers a string argument.
func (a *Args) String(name, help string, opts ...ArgOption) {
	a.register(name, help, "string", false,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(args[0])
			//splitArgs, err := shlex.Split(args[0], true, false )
			//jlog.Error(splitArgs)
			//if err != nil {
			//	return nil, err
			//}
			res[name] = &ArgMapItem{Value: args[0]}
			return args[1:], nil
		},
		opts...,
	)
}

// StringList registers a string list argument.
func (a *Args) StringList(name, help string, opts ...ArgOption) {
	a.register(name, help, "string list", true,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(len(args),args)
			splitArgs, err := shlex.Split(args[0], true, false, ',')
			if err != nil {
				return nil, err
			}
			res[name] = &ArgMapItem{Value: splitArgs}
			args = args[1:]
			return args, nil
		},
		opts...,
	)
}

// Bool registers a bool argument.
func (a *Args) Bool(name, help string, opts ...ArgOption) {
	a.register(name, help, "bool", false,
		func(args []string, res ArgMap) ([]string, error) {
			b, err := strconv.ParseBool(args[0])
			if err != nil {
				return nil, fmt.Errorf("invalid bool value '%s' for argument: %s", args[0], name)
			}

			res[name] = &ArgMapItem{Value: b}
			return args[1:], nil
		},
		opts...,
	)
}

// BoolList registers a bool list argument.
func (a *Args) BoolList(name, help string, opts ...ArgOption) {
	a.register(name, help, "bool list", true,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(len(args),args)
			splitArgs, err2 := shlex.Split(args[0], true, false, ',')
			if err2 != nil {
				return nil, err2
			}
			var (
				err error
				bs  = make([]bool, len(splitArgs))
			)
			for i, a := range splitArgs {
				bs[i], err = strconv.ParseBool(a)
				if err != nil {
					return nil, fmt.Errorf("invalid bool value '%s' for argument: %s", a, name)
				}
			}

			res[name] = &ArgMapItem{Value: bs}
			args = args[1:]
			return args, nil
		},
		opts...,
	)
}

// Int registers an int argument.
func (a *Args) Int(name, help string, opts ...ArgOption) {
	a.register(name, help, "int", false,
		func(args []string, res ArgMap) ([]string, error) {
			i, err := strconv.Atoi(args[0])
			if err != nil {
				return nil, fmt.Errorf("invalid int value '%s' for argument: %s", args[0], name)
			}

			res[name] = &ArgMapItem{Value: i}
			return args[1:], nil
		},
		opts...,
	)
}

// IntList registers an int list argument.
func (a *Args) IntList(name, help string, opts ...ArgOption) {
	a.register(name, help, "int list", true,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(len(args),args)
			splitArgs, err2 := shlex.Split(args[0], true, false, ',')
			if err2 != nil {
				return nil, err2
			}
			var (
				err error
				is  = make([]int, len(splitArgs))
			)
			for i, a := range splitArgs {
				is[i], err = strconv.Atoi(a)
				if err != nil {
					return nil, fmt.Errorf("invalid int value '%s' for argument: %s", a, name)
				}
			}

			res[name] = &ArgMapItem{Value: is}
			//jlog.Error(res)
			args = args[1:]
			return args, nil
		},
		opts...,
	)
}

// Int64 registers an int64 argument.
func (a *Args) Int64(name, help string, opts ...ArgOption) {
	a.register(name, help, "int64", false,
		func(args []string, res ArgMap) ([]string, error) {
			i, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid int64 value '%s' for argument: %s", args[0], name)
			}

			res[name] = &ArgMapItem{Value: i}
			return args[1:], nil
		},
		opts...,
	)
}

// Int64List registers an int64 list argument.
func (a *Args) Int64List(name, help string, opts ...ArgOption) {
	a.register(name, help, "int64 list", true,
		func(args []string, res ArgMap) ([]string, error) {
			var (
				err error
				is  = make([]int64, len(args))
			)
			for i, a := range args {
				is[i], err = strconv.ParseInt(a, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid int64 value '%s' for argument: %s", a, name)
				}
			}

			res[name] = &ArgMapItem{Value: is}
			return []string{}, nil
		},
		opts...,
	)
}

// Uint registers an uint argument.
func (a *Args) Uint(name, help string, opts ...ArgOption) {
	a.register(name, help, "uint", false,
		func(args []string, res ArgMap) ([]string, error) {
			u, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid uint value '%s' for argument: %s", args[0], name)
			}

			res[name] = &ArgMapItem{Value: uint(u)}
			return args[1:], nil
		},
		opts...,
	)
}

// UintList registers an uint list argument.
func (a *Args) UintList(name, help string, opts ...ArgOption) {
	a.register(name, help, "uint list", true,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(len(args),args)
			splitArgs, err2 := shlex.Split(args[0], true, false, ',')
			if err2 != nil {
				return nil, err2
			}
			var (
				err error
				u   uint64
				is  = make([]uint, len(splitArgs))
			)
			for i, a := range splitArgs {
				u, err = strconv.ParseUint(a, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid uint value '%s' for argument: %s", a, name)
				}
				is[i] = uint(u)
			}

			res[name] = &ArgMapItem{Value: is}
			args = args[1:]
			return args, nil
		},
		opts...,
	)
}

// Uint64 registers an uint64 argument.
func (a *Args) Uint64(name, help string, opts ...ArgOption) {
	a.register(name, help, "uint64", false,
		func(args []string, res ArgMap) ([]string, error) {
			u, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid uint64 value '%s' for argument: %s", args[0], name)
			}

			res[name] = &ArgMapItem{Value: u}
			return args[1:], nil
		},
		opts...,
	)
}

// Uint64List registers an uint64 list argument.
func (a *Args) Uint64List(name, help string, opts ...ArgOption) {
	a.register(name, help, "uint64 list", true,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(len(args),args)
			splitArgs, err2 := shlex.Split(args[0], true, false, ',')
			if err2 != nil {
				return nil, err2
			}
			var (
				err error
				us  = make([]uint64, len(splitArgs))
			)
			for i, a := range splitArgs {
				us[i], err = strconv.ParseUint(a, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid uint64 value '%s' for argument: %s", a, name)
				}
			}

			res[name] = &ArgMapItem{Value: us}
			args = args[1:]
			return args, nil
		},
		opts...,
	)
}

// Float64 registers a float64 argument.
func (a *Args) Float64(name, help string, opts ...ArgOption) {
	a.register(name, help, "float64", false,
		func(args []string, res ArgMap) ([]string, error) {
			f, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float64 value '%s' for argument: %s", args[0], name)
			}

			res[name] = &ArgMapItem{Value: f}
			return args[1:], nil
		},
		opts...,
	)
}

// Float64List registers an float64 list argument.
func (a *Args) Float64List(name, help string, opts ...ArgOption) {
	a.register(name, help, "float64 list", true,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(len(args),args)
			splitArgs, err2 := shlex.Split(args[0], true, false, ',')
			if err2 != nil {
				return nil, err2
			}
			var (
				err error
				fs  = make([]float64, len(splitArgs))
			)
			for i, a := range splitArgs {
				fs[i], err = strconv.ParseFloat(a, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid float64 value '%s' for argument: %s", a, name)
				}
			}

			res[name] = &ArgMapItem{Value: fs}
			args = args[1:]
			return args, nil
		},
		opts...,
	)
}

// Duration registers a duration argument.
func (a *Args) Duration(name, help string, opts ...ArgOption) {
	a.register(name, help, "duration", false,
		func(args []string, res ArgMap) ([]string, error) {
			d, err := time.ParseDuration(args[0])
			if err != nil {
				return nil, fmt.Errorf("invalid duration value '%s' for argument: %s", args[0], name)
			}

			res[name] = &ArgMapItem{Value: d}
			return args[1:], nil
		},
		opts...,
	)
}

// DurationList registers an duration list argument.
func (a *Args) DurationList(name, help string, opts ...ArgOption) {
	a.register(name, help, "duration list", true,
		func(args []string, res ArgMap) ([]string, error) {
			//jlog.Error(len(args),args)
			splitArgs, err2 := shlex.Split(args[0], true, false, ',')
			if err2 != nil {
				return nil, err2
			}
			var (
				err error
				ds  = make([]time.Duration, len(splitArgs))
			)
			for i, a := range splitArgs {
				ds[i], err = time.ParseDuration(a)
				if err != nil {
					return nil, fmt.Errorf("invalid duration value '%s' for argument: %s", a, name)
				}
			}

			res[name] = &ArgMapItem{Value: ds}
			args = args[1:]
			return args, nil
		},
		opts...,
	)
}
