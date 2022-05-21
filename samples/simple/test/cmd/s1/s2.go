package s1

import (
	"github.com/chroblert/jishell"
	_ "github.com/chroblert/jishell/samples/simple/test/cmd/s1/s2"
	"github.com/chroblert/jlog"
	"github.com/spf13/viper"
)

var s2Cmd = &jishell.Command{
	Name:      "s2",
	Aliases:   nil,
	Help:      "",
	LongHelp:  "",
	HelpGroup: "",
	Usage:     "",
	Flags:     nil,
	Args:      nil,
	Run: func(c *jishell.Context) error {
		jlog.NInfo("s2")
		return nil
	},
	Completer: nil,
}

func init() {
	var tmpCommands []*jishell.Command
	// 如果该命令有子命令，则进行加载
	if viper.Get("s1_s2Commands") != nil {
		for _, subCmd := range viper.Get("s1_s2Commands").([]*jishell.Command) {
			s2Cmd.AddCommand(subCmd)
		}
	}
	// 被父命令加载
	if viper.Get("s1Commands") == nil {
		tmpCommands = make([]*jishell.Command, 0)
	} else {
		tmpCommands = viper.Get("s1Commands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, s2Cmd)
	viper.Set("s1Commands", tmpCommands)

}
