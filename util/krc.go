package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 酷狗 KRC 逐词歌词解密密钥（16 字节）。
var kugouKRCKey = []byte{
	0x40, 0x47, 0x61, 0x77, 0x5e, 0x32, 0x74, 0x47,
	0x51, 0x36, 0x31, 0x2d, 0xce, 0xd2, 0x6e, 0x69,
}

// DecodeKRC 解密酷狗 KRC blob（base64）→ 明文 KRC 文本。
func DecodeKRC(content string) (string, error) {
	blob, err := base64.StdEncoding.DecodeString(strings.TrimSpace(content))
	if err != nil {
		return "", err
	}
	if len(blob) <= 4 {
		return "", fmt.Errorf("krc blob too short")
	}
	enc := blob[4:] // 跳过 4 字节 "krc1" magic
	plain := make([]byte, len(enc))
	for i, b := range enc {
		plain[i] = b ^ kugouKRCKey[i%len(kugouKRCKey)]
	}
	return inflateZlib(plain)
}

var (
	krcLineRe = regexp.MustCompile(`^\[(\d+),(\d+)\](.*)$`)
	krcWordRe = regexp.MustCompile(`<\d+,\d+,\d+>`)
	krcLangRe = regexp.MustCompile(`^\[language:(.*)\]$`)
)

type krcLanguage struct {
	Content []struct {
		LyricContent [][]string `json:"lyricContent"`
		Type         int        `json:"type"`
	} `json:"content"`
}

// ParseKRC 把明文 KRC 转换为统一 LyricResult：
// 主歌词、翻译（type==1）、罗马音（type==0）均输出为行级 LRC。
func ParseKRC(krc string) *LyricResult {
	lines := strings.Split(krc, "\n")

	type lineTS struct {
		startMS int
		text    string
	}
	var lyricLines []lineTS
	var lang krcLanguage

	for _, raw := range lines {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if m := krcLangRe.FindStringSubmatch(raw); m != nil {
			if decoded, err := base64.StdEncoding.DecodeString(m[1]); err == nil {
				_ = json.Unmarshal(decoded, &lang)
			}
			continue
		}
		m := krcLineRe.FindStringSubmatch(raw)
		if m == nil {
			continue
		}
		startMS, _ := strconv.Atoi(m[1])
		text := krcWordRe.ReplaceAllString(m[3], "")
		text = strings.TrimSpace(text)
		lyricLines = append(lyricLines, lineTS{startMS: startMS, text: text})
	}

	formatTS := func(ms int) string {
		return fmt.Sprintf("[%02d:%02d.%02d]", ms/60000, (ms%60000)/1000, (ms%1000)/10)
	}

	var lrcB, transB, romaB strings.Builder
	// 翻译/罗马音按行索引对齐
	var transLines, romaLines [][]string
	for _, c := range lang.Content {
		switch c.Type {
		case 1:
			transLines = c.LyricContent
		case 0:
			romaLines = c.LyricContent
		}
	}

	for i, ln := range lyricLines {
		ts := formatTS(ln.startMS)
		if ln.text != "" {
			lrcB.WriteString(ts + ln.text + "\n")
		}
		if i < len(transLines) {
			t := strings.TrimSpace(strings.Join(transLines[i], ""))
			if t != "" {
				transB.WriteString(ts + t + "\n")
			}
		}
		if i < len(romaLines) {
			rm := strings.TrimSpace(strings.Join(romaLines[i], ""))
			if rm != "" {
				romaB.WriteString(ts + rm + "\n")
			}
		}
	}

	return &LyricResult{
		Lyric:       strings.TrimRight(lrcB.String(), "\n"),
		Translation: strings.TrimRight(transB.String(), "\n"),
		Roma:        strings.TrimRight(romaB.String(), "\n"),
		Type:        "lrc",
	}
}
