package CDNCheck

import (
	"github.com/chroblert/jishell"
	"github.com/chroblert/jlog"
	"github.com/spf13/viper"
)

var cdnCmd = &jishell.Command{
	Name:      "cdn",
	Aliases:   nil,
	Help:      "检测目标IP是否属于CDN",
	LongHelp:  "",
	HelpGroup: "CDN Check",
	Usage:     "",
	Flags: func(f *jishell.Flags) {
		//f.String("t","dd","ddd","")
		f.StringList("k", "sl", []string{}, "")
	},
	Args: func(a *jishell.Args) {
		a.StringList("t", "")
		a.IntList("i", "")
		a.BoolList("b", "")
		a.DurationList("d", "")
		a.Float64List("f", "")
		a.UintList("ul", "")
		a.Uint64List("u6", "")
		a.Bool("bb", "")
	},
	Run: func(c *jishell.Context) error {
		jlog.Info("test")
		jlog.Info(c.Flags.StringSlice("sl"))
		jlog.Info(len(c.Args.StringList("f")), c.Args.StringList("t"))
		return nil
	},
}

func init() {
	var tmpCommands []*jishell.Command
	if viper.Get("jCommands") == nil {
		tmpCommands = make([]*jishell.Command, 0)
	} else {
		tmpCommands = viper.Get("jCommands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, cdnCmd)
	viper.Set("jCommands", tmpCommands)
}
