package cmd

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jishell"
	_ "github.com/chroblert/jishell/samples/simple/test/cmd/s1"
	"github.com/spf13/viper"
)

var s1Cmd = &jishell.Command{
	Name:      "s1",
	Aliases:   nil,
	Help:      "",
	LongHelp:  "",
	HelpGroup: "",
	Usage:     "",
	Flags: func(f *jishell.Flags) {
		f.StringL("f1", "", "")
		f.StringList("f", "fl2", []string{}, "")
		f.IntL("il", 1, "")
	},
	Args: func(a *jishell.Args) {
		a.String("a1", "")
		a.StringList("al", "")
		a.IntList("il", "")
		a.BoolList("bl", "")
	},
	Run: func(c *jishell.Context) error {
		jlog.Info(c.Flags.String("f1"))
		jlog.Info(c.Args.String("a1"))
		jlog.Info(c.Args.StringList("al"))
		jlog.Info(c.Flags.Int("il"))
		jlog.Info(c.Args.IntList("il"))
		jlog.Info(c.Args.BoolList("bl"))
		return nil
	},
	Completer: nil,
}

func init() {
	var tmpCommands []*jishell.Command
	// 如果该命令有子命令，则进行加载
	if viper.Get("s1Commands") != nil {
		for _, subCmd := range viper.Get("s1Commands").([]*jishell.Command) {
			s1Cmd.AddCommand(subCmd)
		}
	}
	// 被父命令加载
	if viper.Get("jCommands") == nil {
		tmpCommands = make([]*jishell.Command, 0)
	} else {
		tmpCommands = viper.Get("jCommands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, s1Cmd)
	viper.Set("jCommands", tmpCommands)
}
