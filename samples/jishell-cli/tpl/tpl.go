package tpl

func MainTemplate() []byte {
	return []byte(`/*
*/
package main

import "github.com/chroblert/jishell"

func main() {
	jishell.Main(App)
}
`)
}

func AppTemplate() []byte {
	return []byte(`/*
*/
package main
import (
	"github.com/chroblert/jishell"
	{{ if .AppName }}_ "{{ .PkgName }}/{{ .AppName }}/cmd"{{ else }}_ "{{ .PkgName }}/cmd"{{ end }}
	"github.com/spf13/viper"
)
var App = jishell.New(&jishell.Config{
	Name:                  "{{ .AppName2 }}",
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
`)
}

func AddCommandTemplate() []byte {
	return []byte(`package {{ .CmdPkgName}}

import (
	"github.com/chroblert/jishell"
	"github.com/chroblert/jlog"
	"github.com/spf13/viper"
	_ "{{ .CmdImportNamePrefix }}/{{ .CmdName }}"
)

var {{ .CmdName }}Cmd = &jishell.Command{
	Name:      "{{ .CmdName }}",
	Aliases:   nil,
	Help:      "",
	LongHelp:  "",
	HelpGroup: "",
	Usage:     "",
	Flags:     nil,
	Args: nil,
	Run: func(c *jishell.Context) error {
		jlog.NInfo("{{ .CmdName }}")
		return nil
	},
	Completer: nil,
}

func init(){
	var tmpCommands []*jishell.Command
	{{ if .CmdParent }}// 如果该命令有子命令，则进行加载
	if viper.Get("{{ .CmdTplPrefix }}_{{ .CmdName }}Commands") != nil{
		for _,subCmd := range viper.Get("{{ .CmdTplPrefix }}_{{ .CmdName }}Commands").([]*jishell.Command){
			{{ .CmdName }}Cmd.AddCommand(subCmd)
		}
	}
	// 被父命令加载
	if viper.Get("{{ .CmdTplPrefix }}Commands") == nil{
		tmpCommands = make([]*jishell.Command,0)
	}else{
		tmpCommands = viper.Get("{{ .CmdTplPrefix }}Commands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, {{ .CmdName }}Cmd)
	viper.Set("{{ .CmdTplPrefix }}Commands",tmpCommands)
	{{ else }}// 如果该命令有子命令，则进行加载
	if viper.Get("{{ .CmdName }}Commands") != nil{
		for _,subCmd := range viper.Get("{{ .CmdName }}Commands").([]*jishell.Command){
			{{ .CmdName }}Cmd.AddCommand(subCmd)
		}
	}
	// 被父命令加载
	if viper.Get("jCommands") == nil{
		tmpCommands = make([]*jishell.Command,0)
	}else{
		tmpCommands = viper.Get("jCommands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, {{ .CmdName }}Cmd)
	viper.Set("jCommands",tmpCommands){{ end }}
}
`)
}
