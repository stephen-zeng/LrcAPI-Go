package util

import (
	"fmt"
	"log"
	"strings"
)

func ErrorPrinter(err error) {
	if err != nil {
		log.Println(err)
		stackTrace := fmt.Sprintf("%+v", err)
		lines := strings.Split(stackTrace, "\n")
		for _, line := range lines {
			if strings.Contains(line, "urlAPI") {
				log.Println(line)
			}
		}
	}
}
