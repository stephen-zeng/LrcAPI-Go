package command

import "os"

func Arg(args []string) {
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
