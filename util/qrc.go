package util

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"

	qrclib "github.com/jixunmoe-go/qrc"
)

// inflateZlib 尝试 zlib 解压；若失败原样返回错误。
func inflateZlib(data []byte) (string, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// DecodeQRC 将 QQ 返回的 QRC 负载解密为 XML/文本。
// 负载可能是：已解密的 XML/token（以 < 或 [ 开头）、hex 编码的加密数据、
// 或 base64 编码（其内部再是 hex 或加密二进制）。
func DecodeQRC(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty qrc payload")
	}
	if strings.HasPrefix(input, "<") || strings.HasPrefix(input, "[") {
		return input, nil
	}

	// 先尝试 base64：某些接口把 hex/明文再套一层 base64
	if decoded, err := base64.StdEncoding.DecodeString(input); err == nil {
		s := strings.TrimSpace(string(decoded))
		if strings.HasPrefix(s, "<") || strings.HasPrefix(s, "[") {
			return s, nil
		}
		// base64 内层可能是 hex
		if isHexString(s) {
			input = s
		} else if xmlText, err := decryptQRCBytes(decoded); err == nil {
			return xmlText, nil
		}
	}

	// hex 编码的加密数据
	if isHexString(input) {
		raw, err := hex.DecodeString(input)
		if err != nil {
			return "", err
		}
		return decryptQRCBytes(raw)
	}

	return "", fmt.Errorf("unrecognized qrc payload")
}

// decryptQRCBytes 使用 QQ 音乐特有的「buggy DES」（3DES-EDE，非 FIPS 标准）
// 解密并 zlib 解压。委托给 github.com/jixunmoe-go/qrc。
func decryptQRCBytes(raw []byte) (string, error) {
	out, err := qrclib.DecodeQRC(raw)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func isHexString(s string) bool {
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

var qrcLyricContentRe = regexp.MustCompile(`(?s)<Lyric_1\s+[^>]*LyricContent="([^"]*)"`)

// ExtractQRCLyricContent 从解密后的 QRC XML 中取出逐词 token 正文。
func ExtractQRCLyricContent(xmlContent string) string {
	m := qrcLyricContentRe.FindStringSubmatch(xmlContent)
	if m == nil {
		return ""
	}
	return html.UnescapeString(m[1])
}

var (
	// 行头 [start,dur]
	qrcLineHeadRe = regexp.MustCompile(`^\[(\d+),(\d+)\]`)
	// 词标签 (start,dur)
	qrcWordTagRe = regexp.MustCompile(`\(\d+,\d+\)`)
)

// QRCTokenToLRC 把 QRC 逐词 token 正文转换为行级 LRC，
// 去掉词级时间标签，只保留行首 [mm:ss.xx]。
func QRCTokenToLRC(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	var lines []string
	for _, line := range strings.Split(token, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := qrcLineHeadRe.FindStringSubmatch(line)
		if m == nil {
			// 可能是 [ti:]/[ar:] 等元信息或纯文本，跳过时间转换
			continue
		}
		startMS, _ := strconv.Atoi(m[1])
		body := line[len(m[0]):]
		body = qrcWordTagRe.ReplaceAllString(body, "")
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("[%02d:%02d.%02d]%s",
			startMS/60000, (startMS%60000)/1000, (startMS%1000)/10, body))
	}
	return strings.Join(lines, "\n")
}
