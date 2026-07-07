package file

import (
	"lrcAPI/util"
	"testing"
)

// 端到端：写入不完备的日语歌词 → 后台补全逻辑 → 重新读取应变为完备。
// 需要真实模型凭据；未配置则跳过。
func TestCompleteLyricsRoundTrip(t *testing.T) {
	util.LoadEnv("../.env")
	if !util.LLMEnabled() {
		t.Skip("LLM not configured; skipping live round-trip test")
	}

	f := File{FolderName: "__test_artist__ - __test_title__"}
	f.InfoLyric = []InfoLyric{
		{
			ID:     "0",
			Title:  "テスト",
			Artist: "__test_artist__",
			Lyric:  "[00:01.00]こんにちは\n[00:05.00]ありがとう",
			Romaji: "",
			Type:   "lrc",
			Source: "netease",
		},
	}
	if err := f.WriteLyric(); err != nil {
		t.Fatalf("WriteLyric: %v", err)
	}
	t.Cleanup(func() {
		rm := File{FolderName: f.FolderName}
		_ = rm.RemoveLyric()
	})

	// 直接同步调用内部补全逻辑（等价于后台协程执行体）
	if err := completeLyrics(f.FolderName); err != nil {
		t.Fatalf("completeLyrics: %v", err)
	}

	var read File
	read.FolderName = f.FolderName
	if err := read.ReadLyric(); err != nil {
		t.Fatalf("ReadLyric: %v", err)
	}
	if len(read.InfoLyric) != 1 {
		t.Fatalf("expected 1 row, got %d", len(read.InfoLyric))
	}
	got := read.InfoLyric[0]
	t.Logf("lyric after completion:\n%s", got.Lyric)
	t.Logf("romaji after completion:\n%s", got.Romaji)
	if !got.IsComplete {
		t.Errorf("expected IsComplete=true after completion")
	}
	if got.Romaji == "" {
		t.Errorf("expected romaji to be filled")
	}
}
