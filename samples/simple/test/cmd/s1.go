package cmd

import (
	"github.com/chroblert/jishell"
	_ "github.com/chroblert/jishell/samples/simple/test/cmd/s1"
	"github.com/chroblert/jlog"
	"github.com/spf13/viper"
)

var s1Cmd = &jishell.Command{
	Name:      "s1",
	Aliases:   nil,
	Help:      "",
	LongHelp:  "",
	HelpGroup: "",
	Usage:     "",
	Flags:     nil,
	Args:      nil,
	Run: func(c *jishell.Context) error {
		jlog.NInfo("s1")
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
