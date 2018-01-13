package main

import (
	"github.com/alisaifee/confligt/cmd"
)

var version = "master"

func main() {
	cmd.RootCmd.Version = version
	cmd.Execute()
}
