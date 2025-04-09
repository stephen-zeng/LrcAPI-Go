package main

import (
	"lrcAPI/command"
	"lrcAPI/handler"
	"os"
)

func main() {
	command.Arg(os.Args)
	handler.Handler(command.Port, command.Pwd)
	defer os.Exit(0)
}
