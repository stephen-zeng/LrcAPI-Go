package command

import (
	"github.com/common-nighthawk/go-figure"
	"os"
)

func Arg(args []string) {
	figlet := figure.NewFigure("LrcAPI-Go", "", true)
	figlet.Print()
	for index, arg := range args {
		if index == 0 {
			continue
		}
		switch arg {
		case "--port":
			Port = args[index+1]
		case "--pwd":
			Pwd = args[index+1]
		}
	}
	dir, _ := os.Getwd()
	if os.Getenv("PWD") != "" && os.Getenv("PWD") != dir {
		Pwd = os.Getenv("PWD")
	}
}
