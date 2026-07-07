package util

// LyricResult 是各平台歌词获取的统一返回结构。
// Lyric/Translation/Roma 均为「未混合」的独立 LRC 文本，
// 由上层 processor 负责把 Lyric 与 Translation 混合、并把 Roma 放入 romaji 字段。
type LyricResult struct {
	Lyric       string // 主歌词 LRC
	Translation string // 翻译 LRC（可为空）
	Roma        string // 罗马音 LRC（可为空）
	Type        string // 歌词种类：lrc / ttml
}
