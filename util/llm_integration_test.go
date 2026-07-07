package util

import (
	"strings"
	"testing"
)

// 需要真实模型凭据；未配置则跳过。
func TestCompleteLyricLive(t *testing.T) {
	LoadEnv("../.env")
	if !LLMEnabled() {
		t.Skip("LLM not configured; skipping live test")
	}

	// 日语歌词，缺翻译与罗马音
	lyric := "[00:01.00]こんにちは\n[00:05.00]ありがとう\n[00:10.00]さようなら"
	newLyric, newRomaji, changed, err := CompleteLyric(lyric, "")
	if err != nil {
		t.Fatalf("CompleteLyric error: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}
	t.Logf("newLyric:\n%s", newLyric)
	t.Logf("newRomaji:\n%s", newRomaji)

	if !IsLyricComplete(newLyric, newRomaji) {
		t.Errorf("lyric still not complete after CompleteLyric")
	}
	// 校验时间戳保持不变
	for _, ts := range []string{"[00:01.00]", "[00:05.00]", "[00:10.00]"} {
		if !strings.Contains(newLyric, ts) {
			t.Errorf("timestamp %s missing from completed lyric", ts)
		}
		if !strings.Contains(newRomaji, ts) {
			t.Errorf("timestamp %s missing from romaji", ts)
		}
	}
}
