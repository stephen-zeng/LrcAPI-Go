package util

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQRCTokenToLRC(t *testing.T) {
	token := "[1000,2500]He(1000,500)llo(1500,1000) world(2500,1000)\n[ti:x]\n[5000,1000]Bye(5000,1000)"
	got := QRCTokenToLRC(token)
	want := "[00:01.00]Hello world\n[00:05.00]Bye"
	if got != want {
		t.Fatalf("QRCTokenToLRC:\n got: %q\nwant: %q", got, want)
	}
}

func TestQQDecodeLyricBase64LRC(t *testing.T) {
	// trans 字段常见形态：base64 编码的明文 LRC
	plain := "[ti:Test]\n[00:01.00]line one\n[00:02.00]line two"
	enc := base64.StdEncoding.EncodeToString([]byte(plain))
	got := qqDecodeLyric(enc)
	if !strings.Contains(got, "line one") || !strings.Contains(got, "[00:02.00]") {
		t.Fatalf("qqDecodeLyric base64 LRC failed: %q", got)
	}
}

func TestDecodeKRC(t *testing.T) {
	// 构造一个最小 KRC：krc1 magic + xor(zlib(内容))
	inner := "[0,1000]<0,500,0>Hello<500,500,0> World\n"
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	_, _ = zw.Write([]byte(inner))
	_ = zw.Close()
	compressed := zbuf.Bytes()
	blob := append([]byte("krc1"), xorKRC(compressed)...)
	b64 := base64.StdEncoding.EncodeToString(blob)

	plain, err := DecodeKRC(b64)
	if err != nil {
		t.Fatalf("DecodeKRC error: %v", err)
	}
	if !strings.Contains(plain, "Hello") {
		t.Fatalf("DecodeKRC content mismatch: %q", plain)
	}
	res := ParseKRC(plain)
	if !strings.Contains(res.Lyric, "Hello World") || !strings.HasPrefix(res.Lyric, "[00:00.00]") {
		t.Fatalf("ParseKRC lyric mismatch: %q", res.Lyric)
	}
}

// 测试辅助：与 DecodeKRC 逆向的 xor（同一密钥自逆）
func xorKRC(data []byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ kugouKRCKey[i%len(kugouKRCKey)]
	}
	return out
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	content := "# comment\nexport QQ_COOKIE=\"abc=1; def=2\"\nNETEASE_COOKIE=plainval\n\nBAD LINE\n"
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	// 确保环境干净
	_ = os.Unsetenv("QQ_COOKIE")
	_ = os.Unsetenv("NETEASE_COOKIE")
	LoadEnv(p)
	if QQCookie != "abc=1; def=2" {
		t.Fatalf("QQCookie = %q", QQCookie)
	}
	if NeteaseCookie != "plainval" {
		t.Fatalf("NeteaseCookie = %q", NeteaseCookie)
	}
}
