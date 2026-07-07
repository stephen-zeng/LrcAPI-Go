package main

import (
	"lrcAPI/command"
	"lrcAPI/handler"
	"lrcAPI/util"
	"os"
)

func main() {
	// 优先加载 .env（二进制直接运行）/ 容器 environment（docker-compose）中的凭据
	util.LoadEnv("")
	command.Arg(os.Args)
	handler.Handler(command.Port, command.Pwd)
	defer os.Exit(0)
}
