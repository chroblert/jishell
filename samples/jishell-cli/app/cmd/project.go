package cmd

import (
	"fmt"
	"github.com/chroblert/jishell/samples/jishell-cli/tpl"
	"github.com/chroblert/jlog"
	"os"
	"strings"
	"text/template"
)

// Project contains name, license and paths to projects.
type Project struct {
	// v2
	PkgName      string
	Copyright    string
	AbsolutePath string
	Viper        bool
	AppName      string
	AppName2     string
}

type Command struct {
	CmdName             string
	CmdParent           string
	CmdParentHandled    string
	CmdPath             string
	CmdTplPrefix        string
	CmdImportNamePrefix string
	CmdPkgName          string
	*Project
}

func (p *Project) Create() error {
	// check if AbsolutePath exists
	if _, err := os.Stat(p.AbsolutePath); os.IsNotExist(err) {
		// create directory
		if err := os.Mkdir(p.AbsolutePath, 0754); err != nil {
			return err
		}
	}
	// create app.go
	appFile, err := os.Create(fmt.Sprintf("%s/app.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer appFile.Close()
	// create main.go
	mainFile, err := os.Create(fmt.Sprintf("%s/main.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer mainFile.Close()
	appTemplate := template.Must(template.New("app").Parse(string(tpl.AppTemplate())))
	err = appTemplate.Execute(appFile, p)
	if err != nil {
		return err
	}
	mainTemplate := template.Must(template.New("main").Parse(string(tpl.MainTemplate())))
	err = mainTemplate.Execute(mainFile, p)
	if err != nil {
		return err
	}

	// 创建子命令占位目录
	if _, err = os.Stat(fmt.Sprintf("%s/cmd", p.AbsolutePath)); os.IsNotExist(err) {
		err := os.Mkdir(fmt.Sprintf("%s/cmd", p.AbsolutePath), 0751)
		if err != nil {
			jlog.Fatal(err)
		}
	}
	// 创建子命令占位文件
	initFile, err := os.Create(fmt.Sprintf("%s/cmd/init.go", p.AbsolutePath))
	if err != nil {
		return err
	}
	defer initFile.Close()
	initFile.WriteString("package cmd")
	jlog.NInfof("Application %s created at %s\n", p.AppName2, p.AbsolutePath)
	return nil
}

func (c *Command) Create() error {
	//jlog.Error(fmt.Sprintf("%s/%s.go",c.CmdPath,c.CmdName))
	if _, err2 := os.Stat(fmt.Sprintf("%s/%s.go", c.CmdPath, c.CmdName)); !os.IsNotExist(err2) {
		return fmt.Errorf("[!] %s/%s.go文件已存在", strings.ReplaceAll(c.CmdPath, "\\", "/"), c.CmdName)
	}
	cmdFile, err := os.Create(fmt.Sprintf("%s/%s.go", c.CmdPath, c.CmdName))
	if err != nil {
		return err
	}
	defer cmdFile.Close()

	commandTemplate := template.Must(template.New("sub").Parse(string(tpl.AddCommandTemplate())))
	err = commandTemplate.Execute(cmdFile, c)
	if err != nil {
		return err
	}

	// 创建该命令的子命令占位目录
	if _, err = os.Stat(fmt.Sprintf("%s/%s", c.CmdPath, c.CmdName)); os.IsNotExist(err) {
		err := os.Mkdir(fmt.Sprintf("%s/%s", c.CmdPath, c.CmdName), 0751)
		if err != nil {
			jlog.Fatal(err)
		}
	}
	// 创建该命令的子命令占位文件
	cmdSubInitFile, err := os.Create(fmt.Sprintf("%s/%s/init.go", c.CmdPath, c.CmdName))
	if err != nil {
		return err
	}
	cmdSubInitFile.WriteString("package " + c.CmdName)
	defer cmdSubInitFile.Close()
	return nil
}
