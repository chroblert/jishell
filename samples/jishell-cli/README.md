# jishell-cli (jishell generator)

本项目参考[spf13/cobra-cli](https://github.com/spf13/cobra-cli)进行开发，主要用于辅助用户创建jishell应用。

# 使用
1. 下载`jishell-cli`
   - `go get github.com/chroblert/jishell-cli`
2. `go mod init <mod name>`
## 创建app
`jishell-cli init [--package <package name>]`
> - 若不使用flag，则在当前目录创建main.go,app.go及cmd文件
> - 若使用flag，则以`<package name>`为名新建一个目录，在目录中创建main.go,app.go及cmd文件
> 
> 
## 添加命令
`jishell-cli add [-p <parent command path>] <sub command>`
> 新建一个子命令
> - `<parent command>`格式为：自cmd(不包括)开始的命令路径。如：`s1/xx2`
> - 若不传入flag，则在cmd目录下创建一个名为`<sub command>`的.go文件
> - 若传入flag，则在指定的命令路径下创建一个名为`<sub command>`的.go文件

支持`jishell-cli add info/scan/portscan`格式

注：
> 使用-p，则`<sub command>`中不能带有`/`