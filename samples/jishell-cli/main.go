package main

import (
	"github.com/chroblert/jishell"
	"github.com/chroblert/jishell/samples/jishell-cli/app"
	"github.com/chroblert/jlog"
)

func main() {
	jlog.SetStoreToFile(false)
	jishell.Main(app.App)
}
