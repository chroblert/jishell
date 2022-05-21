package CDNCheck

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jishell"
	"github.com/spf13/viper"
)

var cdnChkDomainCmd = &jishell.Command{
	Name:      "cdnChkD",
	Help:      "检测目标域名是否使用了CDN",
	LongHelp:  "",
	HelpGroup: "CDN Check",
	Usage:     "",
	Flags: func(f *jishell.Flags) {
		f.Bool("b", "boolf", true, "")
	},
	Args: func(a *jishell.Args) {
		a.String("host", "hostname.eg: www.test.com")
		a.Bool("verbose", "")
	},
	Run: func(c *jishell.Context) error {
		jlog.Info("boolf:", c.Flags.Bool("boolf"))
		jlog.Info("host:", c.Args.String("host"))
		jlog.Info("verbose:", c.Args.Bool("verbose"))
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
	tmpCommands = append(tmpCommands, cdnChkDomainCmd)
	viper.Set("jCommands", tmpCommands)
}
