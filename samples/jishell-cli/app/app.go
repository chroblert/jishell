package app

import (
	"github.com/chroblert/jishell"
	_ "github.com/chroblert/jishell/samples/jishell-cli/app/cmd"
	"github.com/spf13/viper"
)

var App = jishell.New(&jishell.Config{
	Name:                  "jishell",
	Description:           "",
	Flags:                 nil,
	HistoryFile:           "",
	HistoryLimit:          0,
	NoColor:               false,
	VimMode:               false,
	Prompt:                "",
	PromptColor:           nil,
	MultiPrompt:           "",
	MultiPromptColor:      nil,
	ASCIILogoColor:        nil,
	ErrorColor:            nil,
	HelpHeadlineUnderline: false,
	HelpSubCommands:       false,
	HelpHeadlineColor:     nil,
})

func init() {
	App.OnInit(func(a *jishell.App, flags jishell.FlagMap) error {
		if viper.Get("jCommands") != nil {
			for _, v := range viper.Get("jCommands").([]*jishell.Command) {
				a.AddCommand(v)
			}
		}
		return nil
	})
}
