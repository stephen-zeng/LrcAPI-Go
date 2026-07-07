package util

import (
	"regexp"
	"strings"
	"unicode"
)

// LyricLine 表示一行带时间戳的歌词。
type LyricLine struct {
	Timestamp string // 该行的第一个时间戳，如 "[00:12.34]"
	Text      string // 去除时间戳后的文本（已 Trim）
}

// 时间戳形如 [mm:ss.xx] 或 [mm:ss.xxx]，分隔符可为 . 或 :
var lrcTimeTagRe = regexp.MustCompile(`\[\d{1,2}:\d{2}[.:]\d{2,3}\]`)

// ParseLrc 把一段 LRC 文本解析成有序的行序列。
// 元数据标签（如 [ti:]/[ar:]）以及没有时间戳的行会被忽略。
// 若一行包含多个时间戳（重复段落），会为每个时间戳各产出一行。
func ParseLrc(lrc string) []LyricLine {
	var lines []LyricLine
	for _, raw := range strings.Split(lrc, "\n") {
		tags := lrcTimeTagRe.FindAllStringIndex(raw, -1)
		if len(tags) == 0 {
			continue
		}
		// 文本是最后一个时间戳之后的内容
		text := strings.TrimSpace(raw[tags[len(tags)-1][1]:])
		for _, t := range tags {
			lines = append(lines, LyricLine{
				Timestamp: raw[t[0]:t[1]],
				Text:      text,
			})
		}
	}
	return lines
}

// originalAndTranslation 把（可能已混合翻译的）歌词按时间戳分组，
// 返回每个时间戳的「原文」行与「译文」行。
// 混合格式为同一时间戳连续出现两次：第一次为原文，第二次为译文。
func originalAndTranslation(lrc string) (original []LyricLine, translation []LyricLine) {
	lines := ParseLrc(lrc)
	seen := make(map[string]int)
	for _, ln := range lines {
		seen[ln.Timestamp]++
		if seen[ln.Timestamp] == 1 {
			original = append(original, ln)
		} else if seen[ln.Timestamp] == 2 {
			translation = append(translation, ln)
		}
	}
	return original, translation
}

// LyricLang 表示歌词主体语言。
type LyricLang int

const (
	LangUnknown  LyricLang = iota // 纯音乐 / 无可判断文本
	LangChinese                   // 中文
	LangJapanese                  // 日语
	LangKorean                    // 韩语
	LangOther                     // 其它外语（英语等拉丁语系）
)

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func hasKana(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
			return true
		}
	}
	return false
}

func hasHangul(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

func hasLatinLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && r < unicode.MaxLatin1 {
			return true
		}
	}
	return false
}

// detectLang 根据原文文本判断主体语言。
func detectLang(text string) LyricLang {
	switch {
	case hasKana(text):
		return LangJapanese
	case hasHangul(text):
		return LangKorean
	case hasHan(text):
		// 有汉字但无假名/谚文，按中文处理（无需翻译）
		return LangChinese
	case hasLatinLetter(text):
		return LangOther
	default:
		return LangUnknown
	}
}

func joinOriginalText(lines []LyricLine) string {
	var sb strings.Builder
	for _, ln := range lines {
		sb.WriteString(ln.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// HasLyricContent 判断歌词是否含有效内容：至少存在一行「带时间戳且文本非空」的歌词。
// 空串、仅有元数据标签、只有时间戳而无文字等「数据不足或异常」的情况均返回 false，
// 用于在来源层过滤掉无用候选（例如网易云只上传了罗马音、主歌词为空的条目）。
func HasLyricContent(lyric string) bool {
	for _, ln := range ParseLrc(lyric) {
		if strings.TrimSpace(ln.Text) != "" {
			return true
		}
	}
	return false
}

// LyricLang 判断歌词（可能已混合翻译）原文的主体语言。
func LyricLanguage(lyric string) LyricLang {
	original, _ := originalAndTranslation(lyric)
	return detectLang(joinOriginalText(original))
}

// IsLyricComplete 判断一条歌词信息是否完备：
//   - 外语（含日/韩）歌词必须带有中文翻译；
//   - 日语 / 韩语歌词还必须带有罗马音；
//   - 中文歌词、纯音乐视为完备。
//
// lyric 为（可能已混合翻译的）主歌词，romaji 为罗马音字段。
func IsLyricComplete(lyric, romaji string) bool {
	original, translation := originalAndTranslation(lyric)
	lang := detectLang(joinOriginalText(original))

	switch lang {
	case LangChinese, LangUnknown:
		return true
	case LangOther:
		return hasChineseTranslation(translation)
	case LangJapanese, LangKorean:
		return hasChineseTranslation(translation) && hasRomaji(romaji)
	default:
		return true
	}
}

// hasChineseTranslation 判断译文行中是否存在中文内容。
func hasChineseTranslation(translation []LyricLine) bool {
	for _, ln := range translation {
		if hasHan(ln.Text) {
			return true
		}
	}
	return false
}

// hasRomaji 判断罗马音字段是否含有效内容（存在拉丁字母）。
func hasRomaji(romaji string) bool {
	return hasLatinLetter(romaji)
}
