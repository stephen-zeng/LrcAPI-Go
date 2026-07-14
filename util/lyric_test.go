package util

import "testing"

func TestLrcTranslationBlenderSeparatesFinalLines(t *testing.T) {
	original := "[00:01.00]一行目\n[00:02.00]電子の海の８小節。"
	translation := "[00:01.00]第一句\n[00:02.00]电子海洋的8小节"

	got := LrcTranslationBlender(original, translation)
	want := "[00:01.00]一行目\n[00:01.00]第一句\n[00:02.00]電子の海の８小節。\n[00:02.00]电子海洋的8小节"
	if got != want {
		t.Fatalf("unexpected blended lyric:\n got: %q\nwant: %q", got, want)
	}
}
