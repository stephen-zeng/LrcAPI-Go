package util

import "testing"

func TestDetectLang(t *testing.T) {
	cases := []struct {
		text string
		want LyricLang
	}{
		{"こんにちは 世界", LangJapanese},
		{"안녕하세요", LangKorean},
		{"你好世界", LangChinese},
		{"Hello world", LangOther},
		{"~~~ +++", LangUnknown},
		{"漢字だけ kanji with かな", LangJapanese},
	}
	for _, c := range cases {
		if got := detectLang(c.text); got != c.want {
			t.Errorf("detectLang(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}

func TestParseLrcAndSplit(t *testing.T) {
	lrc := "[ti:test]\n[00:01.00]こんにちは\n[00:01.00]你好\n[00:02.50]世界\n[00:02.50]世界（中）"
	original, translation := originalAndTranslation(lrc)
	if len(original) != 2 {
		t.Fatalf("expected 2 original lines, got %d", len(original))
	}
	if len(translation) != 2 {
		t.Fatalf("expected 2 translation lines, got %d", len(translation))
	}
	if original[0].Text != "こんにちは" || original[0].Timestamp != "[00:01.00]" {
		t.Errorf("unexpected original[0]: %+v", original[0])
	}
	if translation[0].Text != "你好" {
		t.Errorf("unexpected translation[0]: %+v", translation[0])
	}
}

func TestIsLyricComplete(t *testing.T) {
	cases := []struct {
		name   string
		lyric  string
		romaji string
		want   bool
	}{
		{
			name:  "chinese only is complete",
			lyric: "[00:01.00]你好世界",
			want:  true,
		},
		{
			name:  "instrumental is complete",
			lyric: "[00:01.00]~~~",
			want:  true,
		},
		{
			name:  "english without translation is incomplete",
			lyric: "[00:01.00]Hello world",
			want:  false,
		},
		{
			name:  "english with chinese translation is complete",
			lyric: "[00:01.00]Hello world\n[00:01.00]你好世界",
			want:  true,
		},
		{
			name:  "japanese missing translation and romaji is incomplete",
			lyric: "[00:01.00]こんにちは",
			want:  false,
		},
		{
			name:  "japanese with translation but no romaji is incomplete",
			lyric: "[00:01.00]こんにちは\n[00:01.00]你好",
			want:  false,
		},
		{
			name:   "japanese with translation and romaji is complete",
			lyric:  "[00:01.00]こんにちは\n[00:01.00]你好",
			romaji: "[00:01.00]konnichiwa",
			want:   true,
		},
		{
			name:   "korean with translation and romaji is complete",
			lyric:  "[00:01.00]안녕\n[00:01.00]你好",
			romaji: "[00:01.00]annyeong",
			want:   true,
		},
	}
	for _, c := range cases {
		if got := IsLyricComplete(c.lyric, c.romaji); got != c.want {
			t.Errorf("%s: IsLyricComplete = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestLlmTransformValidation(t *testing.T) {
	// buildBlended 正确交错原文与译文
	original := []LyricLine{
		{Timestamp: "[00:01.00]", Text: "Hello"},
		{Timestamp: "[00:02.00]", Text: "World"},
	}
	trans := map[int]string{0: "你好", 1: "世界"}
	got := buildBlended(original, trans)
	want := "[00:01.00]Hello\n[00:01.00]你好\n[00:02.00]World\n[00:02.00]世界"
	if got != want {
		t.Errorf("buildBlended =\n%q\nwant\n%q", got, want)
	}
}

func TestHasLyricContent(t *testing.T) {
	cases := []struct {
		name  string
		lyric string
		want  bool
	}{
		{"empty", "", false},
		{"metadata only", "[ti:song]\n[ar:artist]", false},
		{"timestamps without text", "[00:01.00]\n[00:02.00]   ", false},
		{"real lyric", "[00:01.00]hello", true},
		{"real lyric among blanks", "[00:01.00]\n[00:02.00]world", true},
	}
	for _, c := range cases {
		if got := HasLyricContent(c.lyric); got != c.want {
			t.Errorf("%s: HasLyricContent = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestStripCodeFence(t *testing.T) {
	in := "```\nline1\nline2\n```"
	if got := stripCodeFence(in); got != "line1\nline2" {
		t.Errorf("stripCodeFence = %q", got)
	}
	in2 := "plain\ntext"
	if got := stripCodeFence(in2); got != "plain\ntext" {
		t.Errorf("stripCodeFence unchanged = %q", got)
	}
}
