package cmd

import (
	"github.com/fatih/color"
	"log"
	"os"
)

func yellow(value interface{}) string {
	return color.New(color.FgYellow).SprintFunc()(value)
}

func red(value interface{}) string {
	return color.New(color.FgRed).SprintFunc()(value)
}
func green(value interface{}) string {
	return color.New(color.FgGreen).SprintFunc()(value)
}
func blue(value interface{}) string {
	return color.New(color.FgBlue).SprintFunc()(value)
}
func cyan(value interface{}) string {
	return color.New(color.FgCyan).SprintFunc()(value)
}

func boolColor(value interface{}, condition bool, colors ...color.Attribute) string {
	var r, g color.Attribute
	if r, g = color.FgRed, color.FgGreen; len(colors) == 2 {
		g = colors[0]
		r = colors[1]
	}
	if condition {
		return color.New(g).SprintFunc()(value)
	} else {
		return color.New(r).SprintFunc()(value)
	}
}

var L *log.Logger
var V bool

func init() {
	L = log.New(os.Stderr, "", 0)
}
