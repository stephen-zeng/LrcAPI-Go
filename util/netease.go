package util

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	neteaseLrcURLTemplate   string = `https://music.163.com/api/song/lyric?id=%s&lv=1&tv=1`
	neteaseSerachIDTemplate string = `https://music.163.com/api/search/get/web?csrf_token=hlpretag=&hlposttag=&s=%s&type=1&offset=0&total=true&limit=100`
)

type neteaseLrcResponse struct {
	Lrc struct {
		Lyric string `json:"lyric"`
	} `json:"lrc"`
	TLyric struct {
		Lyric string `json:"lyric"`
	} `json:"tlyric"`
}

type neteaseSerachIDResponse struct {
	Result struct {
		Songs []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"songs"`
	} `json:"result"`
}

// 通过ID查找歌词
// 返回歌词，翻译歌词以及错误
// 返回的歌词都是未转译的
func NeteaseGetLyric(ID string) (string, string, error) {
	url := fmt.Sprintf(neteaseLrcURLTemplate, ID)
	req, _ := http.NewRequest("GET", url, nil)
	SetHeader(req)
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("Origin", "https://music.163.com/")
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return "", "", errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return "", "", errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var response neteaseLrcResponse
	_ = json.Unmarshal(body, &response)
	return response.Lrc.Lyric, response.TLyric.Lyric, nil
}

// 搜索
// 给标题和作者，返回ID序列，曲名序列，作者序列的指针
func NeteaseGetMusic(Title, Artist string) (*[]string, *[]string, *[]string, error) {
	url := strings.ReplaceAll(fmt.Sprintf(neteaseSerachIDTemplate, url.PathEscape(Title+` `+Artist)), ` `, "%20")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", UA)
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("Origin", "https://music.163.com/")
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, nil, nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, nil, nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var response neteaseSerachIDResponse
	_ = json.Unmarshal(body, &response)
	ids := make([]string, len(response.Result.Songs))
	names := make([]string, len(response.Result.Songs))
	artists := make([]string, len(response.Result.Songs))
	for index, song := range response.Result.Songs {
		ids[index] = strconv.Itoa(song.ID)
		names[index] = song.Name
		artists[index] = song.Artists[0].Name
	}
	return &ids, &names, &artists, nil
}
