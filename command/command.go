package command

import (
	"github.com/common-nighthawk/go-figure"
	"os"
)

func Arg(args []string) {
	figlet := figure.NewFigure("LrcAPI-Go", "", true)
	figlet.Print()

	// 1) 环境变量（含 .env 注入）优先级低于命令行参数
	if v := os.Getenv("PORT"); v != "" {
		Port = v
	}
	if v := os.Getenv("LRCAPI_PORT"); v != "" {
		Port = v
	}
	// 兼容旧行为：docker 通过 -e PWD=xxx 传密码。
	// PWD 在多数 shell 中默认是工作目录，因此仅当它与工作目录不同才视为密码。
	dir, _ := os.Getwd()
	if p := os.Getenv("PWD"); p != "" && p != dir {
		Pwd = p
	}
	// 更清晰、无歧义的密码变量（推荐在 .env 中使用）
	if v := os.Getenv("LRCAPI_PWD"); v != "" {
		Pwd = v
	}

	// 2) 命令行参数最高优先级
	for index, arg := range args {
		if index == 0 {
			continue
		}
		switch arg {
		case "--port":
			if index+1 < len(args) {
				Port = args[index+1]
			}
		case "--pwd":
			if index+1 < len(args) {
				Pwd = args[index+1]
			}
		}
	}
}
