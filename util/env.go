package util

import (
	"bufio"
	"os"
	"strings"
)

// 各平台可选凭据。留空表示以匿名方式访问对应平台。
var (
	// 网易云 MUSIC_U / MUSIC_A Cookie，提升命中率
	NeteaseCookie string
	// QQ 音乐 Cookie（含 uin/qqmusic_key 等），可选
	QQCookie string
	// 酷狗 Cookie，主要用于歌词下载，可选
	KugouCookie string
	// Apple Music：Developer Token（Bearer）与 media-user-token
	AppleDeveloperToken string
	AppleMediaUserToken string
	// Spotify sp_dc Cookie（用于派生 access token）
	SpotifySpDc string
	// 汽水音乐 Cookie，可选
	SodaCookie string
	// 哔哩哔哩 Cookie（SESSDATA 等），字幕接口通常需要
	BilibiliCookie string
	// YouTube Music Cookie，可选
	YoutubeMusicCookie string
)

// 大模型（OpenAI 兼容接口）配置，用于补齐翻译与罗马音。
// 留空则关闭自动补全功能。
var (
	// OpenAI 兼容接口的 API Key
	OpenAIAPIKey string
	// OpenAI 兼容接口的 Base URL（不需要含 /v1 等路径），如 https://api.deepseek.com
	OpenAIBaseURL string
	// 使用的模型名，如 deepseek-v4-flash
	OpenAIModel string
)

// LoadEnv 读取 .env 文件（若存在）并把其中的键值对写入进程环境变量，
// 已存在的环境变量不会被覆盖。随后把平台凭据读取到包级变量。
// 这样二进制直接运行时可用 .env，容器部署时可用 docker-compose 的 environment。
func LoadEnv(path string) {
	if path == "" {
		path = ".env"
	}
	loadDotEnvFile(path)

	NeteaseCookie = os.Getenv("NETEASE_COOKIE")
	QQCookie = os.Getenv("QQ_COOKIE")
	KugouCookie = os.Getenv("KUGOU_COOKIE")
	AppleDeveloperToken = os.Getenv("APPLE_DEVELOPER_TOKEN")
	AppleMediaUserToken = os.Getenv("APPLE_MEDIA_USER_TOKEN")
	SpotifySpDc = os.Getenv("SPOTIFY_SP_DC")
	SodaCookie = os.Getenv("SODA_COOKIE")
	BilibiliCookie = os.Getenv("BILIBILI_COOKIE")
	YoutubeMusicCookie = os.Getenv("YOUTUBE_MUSIC_COOKIE")

	OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	OpenAIBaseURL = strings.TrimRight(os.Getenv("OPENAI_BASE_URL"), "/")
	OpenAIModel = os.Getenv("OPENAI_MODEL")
}

// loadDotEnvFile 解析形如 KEY=VALUE 的 .env 文件。
// 支持 # 注释、空行、可选的 export 前缀以及引号包裹的值。
func loadDotEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // 没有 .env 是正常情况
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.Index(line, "=")
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}
