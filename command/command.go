package command

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
}
