package util

import (
	"fmt"
	"log"
	"net/http"
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

// 将请求伪装成浏览器
func SetHeader(req *http.Request) {
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("priority", "u=0, i")
	req.Header.Set("sec-ch-ua", `"Chromium";v="135", "Not-A.Brand";v="8"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", UA)
}
