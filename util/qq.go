package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	// 官方逐词歌词接口（QRC）与搜索都走 musicu.fcg
	qqMusicuURL string = `https://u.y.qq.com/cgi-bin/musicu.fcg?format=json&inCharset=utf8&outCharset=utf8`
	// 官方 LRC 回退接口
	qqLrcFallbackTemplate string = `https://c.y.qq.com/lyric/fcgi-bin/fcg_query_lyric_new.fcg?songmid=%s&g_tk=5381&format=json&inCharset=utf8&outCharset=utf-8&platform=yqq`
	qqUA                  string = `QQMusic/14090508 (android 12)`
)

// 现代 musicu.fcg 搜索模块响应
type qqSearchResponse struct {
	Req1 struct {
		Data struct {
			Body struct {
				Song struct {
					List []qqSongInfo `json:"list"`
				} `json:"song"`
			} `json:"body"`
		} `json:"data"`
	} `json:"req_1"`
}

type qqSongInfo struct {
	ID     int    `json:"id"`
	Mid    string `json:"mid"`
	Title  string `json:"title"`
	Singer []struct {
		Name string `json:"name"`
	} `json:"singer"`
}

func qqConnectSinger(song *qqSongInfo) string {
	var names []string
	for _, singer := range song.Singer {
		if singer.Name != "" {
			names = append(names, singer.Name)
		}
	}
	return strings.Join(names, ", ")
}

// 给标题（可含歌手），返回 id, mid, title, artist 的切片指针。
// 使用官方 musicu.fcg 搜索模块（旧的 client_search_cp 已陆续下线/返回 500）。
func QQGetMusic(keyword string) (*[]string, *[]string, *[]string, *[]string, error) {
	reqBody := map[string]any{
		"req_1": map[string]any{
			"method": "DoSearchForQQMusicDesktop",
			"module": "music.search.SearchCgiService",
			"param": map[string]any{
				"num_per_page": 20,
				"page_num":     1,
				"query":        keyword,
				"search_type":  0,
			},
		},
	}
	payload, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", qqMusicuURL, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", qqUA)
	req.Header.Set("Referer", "https://y.qq.com/")
	req.Header.Set("Origin", "https://y.qq.com")
	if QQCookie != "" {
		req.Header.Set("Cookie", QQCookie)
	}
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, nil, nil, nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, nil, nil, nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var response qqSearchResponse
	_ = json.Unmarshal(body, &response)
	songList := response.Req1.Data.Body.Song.List
	var ids, mids, names, artists []string
	for _, songInfo := range songList {
		if songInfo.Mid == "" {
			continue
		}
		ids = append(ids, strconv.Itoa(songInfo.ID))
		mids = append(mids, songInfo.Mid)
		names = append(names, songInfo.Title)
		artists = append(artists, qqConnectSinger(&songInfo))
	}
	return &ids, &mids, &names, &artists, nil
}

// ---- 逐词歌词（QRC）接口 ----

type qqMusicuRequest struct {
	Comm map[string]any `json:"comm"`
	Req1 struct {
		Method string         `json:"method"`
		Module string         `json:"module"`
		Param  map[string]any `json:"param"`
	} `json:"req_1"`
}

type qqMusicuResponse struct {
	Req1 struct {
		Code int `json:"code"`
		Data struct {
			Lyric string `json:"lyric"`
			Trans string `json:"trans"`
			Roma  string `json:"roma"`
			Qrc   int    `json:"qrc"`
		} `json:"data"`
	} `json:"req_1"`
}

func qqBuildMusicuBody(songMID string) ([]byte, error) {
	var reqBody qqMusicuRequest
	reqBody.Comm = map[string]any{
		"_channelid":   "0",
		"_os_version":  "6.2.9200-2",
		"authst":       "",
		"ct":           "19",
		"cv":           "1873",
		"patch":        "118",
		"tmeAppID":     "qqmusic",
		"tmeLoginType": 2,
		"uin":          "0",
		"wid":          "0",
	}
	reqBody.Req1.Method = "GetPlayLyricInfo"
	reqBody.Req1.Module = "music.musichallSong.PlayLyricInfo"
	reqBody.Req1.Param = map[string]any{
		"songMID": songMID,
		"qrc":     1,
		"roma":    1,
		"trans":   1,
	}
	return json.Marshal(reqBody)
}

// QQGetLyricByMID 优先调用官方 QRC 接口，失败时回退官方 LRC 接口。
func QQGetLyricByMID(mid string) (*LyricResult, error) {
	if res, err := qqGetQRC(mid); err == nil && res != nil && strings.TrimSpace(res.Lyric) != "" {
		return res, nil
	}
	return qqGetLRCFallback(mid)
}

func qqGetQRC(mid string) (*LyricResult, error) {
	payload, err := qqBuildMusicuBody(mid)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req, _ := http.NewRequest("POST", qqMusicuURL, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", qqUA)
	req.Header.Set("Referer", "https://y.qq.com/")
	req.Header.Set("Origin", "https://y.qq.com")
	if QQCookie != "" {
		req.Header.Set("Cookie", QQCookie)
	}
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	body, _ := io.ReadAll(resp.Body)
	var response qqMusicuResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.WithStack(err)
	}
	data := response.Req1.Data
	lyric := qqDecodeLyric(data.Lyric)
	if strings.TrimSpace(lyric) == "" {
		return nil, errors.New("empty qrc lyric")
	}
	return &LyricResult{
		Lyric:       lyric,
		Translation: qqDecodeLyric(data.Trans),
		Roma:        qqDecodeLyric(data.Roma),
		Type:        "lrc",
	}, nil
}

type qqLrcFallbackResponse struct {
	RetCode int    `json:"retcode"`
	Lyric   string `json:"lyric"`
	Trans   string `json:"trans"`
}

func qqGetLRCFallback(mid string) (*LyricResult, error) {
	u := fmt.Sprintf(qqLrcFallbackTemplate, mid)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", qqUA)
	req.Header.Set("Referer", "https://y.qq.com/portal/player.html")
	if QQCookie != "" {
		req.Header.Set("Cookie", QQCookie)
	}
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	body, _ := io.ReadAll(resp.Body)
	// 该接口可能返回 JSONP，需要剥离外层
	jsonStr := extractJSONFromJSONP(string(body))
	if jsonStr == "" {
		jsonStr = string(body)
	}
	var response qqLrcFallbackResponse
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, errors.WithStack(err)
	}
	return &LyricResult{
		Lyric:       qqDecodeLyric(response.Lyric),
		Translation: qqDecodeLyric(response.Trans),
		Type:        "lrc",
	}, nil
}

// 提取JSON，为QQ音乐准备（部分接口返回 JSONP 包裹）
func extractJSONFromJSONP(jsonp string) string {
	start := strings.Index(jsonp, "(")
	end := strings.LastIndex(jsonp, ")")
	if start >= 0 && end > start {
		return jsonp[start+1 : end]
	}
	return ""
}

var (
	qqLRCTagRe = regexp.MustCompile(`\[\d+:\d+`)
	qqQRCTagRe = regexp.MustCompile(`(?m)^\[\d+,\d+\]`)
)

// qqDecodeLyric 把 QQ 各接口返回的歌词字段规整为行级 LRC。
// 支持：明文 LRC、base64(LRC)、QRC hex/加密、QRC XML。
func qqDecodeLyric(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	decoded := raw
	if !strings.HasPrefix(raw, "[") && !strings.HasPrefix(raw, "<") {
		if d, err := DecodeQRC(raw); err == nil && strings.TrimSpace(d) != "" {
			decoded = strings.TrimSpace(d)
		} else if b, err := base64.StdEncoding.DecodeString(raw); err == nil {
			decoded = strings.TrimSpace(string(b))
		}
	}

	// XML 形式：抽取逐词 token 正文
	if strings.Contains(decoded, "<Lyric_") {
		if content := ExtractQRCLyricContent(decoded); content != "" {
			decoded = strings.TrimSpace(content)
		}
	}

	// QRC 逐词 token → 行级 LRC
	if qqQRCTagRe.MatchString(decoded) {
		if lrc := QRCTokenToLRC(decoded); lrc != "" {
			return lrc
		}
	}

	// 已是 LRC
	if qqLRCTagRe.MatchString(decoded) {
		return decoded
	}
	return ""
}
