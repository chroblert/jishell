package main

import (
	"github.com/chroblert/jishell"
	"github.com/chroblert/jishell/samples/cmd"
)

func main() {
	jishell.Main(cmd.App)
}
