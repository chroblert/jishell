package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/chroblert/jishell"
	"github.com/chroblert/jlog"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var initCmd = &jishell.Command{
	Name:      "init",
	Aliases:   []string{"initialize", "create"},
	Help:      "Initialize a jishell Application",
	LongHelp:  "",
	HelpGroup: "",
	Usage:     "init [path]",
	Flags: func(f *jishell.Flags) {
		f.StringL("package", "", "package name")
	},
	Args: func(a *jishell.Args) {
		//a.String("package","package name")
	},
	Run: func(c *jishell.Context) error {
		packageName := c.Flags.String("package")
		//if len(packageName) == 0{
		//	return fmt.Errorf("请输入符合格式的package")
		//}
		goGet("github.com/chroblert/jishell")
		goGet("github.com/chroblert/jlog")
		goGet("github.com/spf13/viper")
		initializeProject(packageName)
		return nil
	},
	Completer: nil,
}

func init() {
	var tmpCommands []*jishell.Command
	if viper.Get("jCommands") == nil {
		tmpCommands = make([]*jishell.Command, 0)
	} else {
		tmpCommands = viper.Get("jCommands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, initCmd)
	viper.Set("jCommands", tmpCommands)
}

func initializeProject(appName string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	modName := getModImportPath()
	appName2 := appName
	if len(appName) > 0 {
		if appName != "." {
			wd = fmt.Sprintf("%s/%s", wd, appName)
		}
	} else {
		if strings.ContainsRune(modName, '/') {
			appName2 = strings.Split(modName, "/")[len(strings.Split(modName, "/"))-1]
		} else if len(modName) > 0 {
			appName2 = modName
		} else {
			appName2 = "jishell"
		}
	}

	//jlog.Info(modName)
	project := &Project{
		AbsolutePath: wd,
		PkgName:      modName,
		//Legal:        getLicense(),
		//Copyright:    copyrightLine(),
		Viper:    viper.GetBool("useViper"),
		AppName:  appName,
		AppName2: appName2,
	}
	//jlog.Info(*project)
	if err := project.Create(); err != nil {
		return "", err
	}

	return project.AbsolutePath, nil
}
func getModImportPath() string {
	mod, cd := parseModInfo()
	//jlog.Info(mod)
	//jlog.Info(cd)
	return path.Join(mod.Path, fileToURL(strings.TrimPrefix(cd.Dir, mod.Dir)))
}

func fileToURL(in string) string {
	i := strings.Split(in, string(filepath.Separator))
	return path.Join(i...)
}

func parseModInfo() (Mod, CurDir) {
	var mod Mod
	var dir CurDir

	m := modInfoJSON("-m")
	if err := json.Unmarshal(m, &mod); err != nil {
		jlog.Fatal(err)
	}

	// Unsure why, but if no module is present Path is set to this string.
	if mod.Path == "command-line-arguments" {
		jlog.Fatal("Please run `go mod init <MODNAME>` before `cobra-cli init`")
	}

	e := modInfoJSON("-e")
	if err := json.Unmarshal(e, &dir); err != nil {
		jlog.Fatal(err)
	}

	return mod, dir
}

type Mod struct {
	Path, Dir, GoMod string
}

type CurDir struct {
	Dir string
}

func goGet(mod string) error {
	return exec.Command("go", "get", mod).Run()
}

func modInfoJSON(args ...string) []byte {
	cmdArgs := append([]string{"list", "-json"}, args...)
	out, err := exec.Command("go", cmdArgs...).Output()
	if err != nil {
		jlog.Fatal(err)
	}
	return out
}
