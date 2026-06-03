package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func warn(format string, a ...any) {
	color.New(color.FgYellow).Fprintf(os.Stderr, "! "+format+"\n", a...)
}

func info(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}

func success(format string, a ...any) {
	color.New(color.FgGreen, color.Bold).Printf(format+"\n", a...)
}
