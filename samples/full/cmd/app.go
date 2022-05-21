package cmd

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jishell"
	_ "github.com/chroblert/jishell/samples/full/cmd/CDNCheck"
	"github.com/spf13/viper"
)

var App = jishell.New(&jishell.Config{
	Name:        "CheckTest",
	Description: "",
	Flags: func(f *jishell.Flags) {
		f.BoolL("verbose", false, "")
	},
	//CurrentCmdStr: "CheckTest",
})

func init() {
	App.SetPrintASCIILogo(func(a *jishell.App) {
		jlog.Warn("=============================")
		jlog.Warn("         CDN Check          ")
		jlog.Warn("=============================")
	})
	App.OnInit(func(a *jishell.App, flags jishell.FlagMap) error {
		if flags.Bool("verbose") {
			jlog.Info("verbose")
		}
		// load commands
		if viper.Get("jCommands") != nil {
			for _, v := range viper.Get("jCommands").([]*jishell.Command) {
				a.AddCommand(v)
			}
		}
		return nil
	})
}
