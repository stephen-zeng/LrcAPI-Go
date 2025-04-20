package util

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	qqSearchIDTemplate string = `https://c.y.qq.com/soso/fcgi-bin/client_search_cp?g_tk=5381&p=1&n=20&w=%s&format=json&loginUin=0&hostUin=0&inCharset=utf8&outCharset=utf-8¬ice=0&platform=yqq&needNewCode=0&remoteplace=txt.yqq.song&t=0&aggr=1&cr=1&catZhida=1&flag_qc=0`
	qqLrcURLTemplate   string = `https://api.xingzhige.com/API/lyrc/?id=%s&mid=%s`
)

func init() {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	GlobalHTTPClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

type qqSerachIDResponse struct {
	Data struct {
		Song struct {
			List []qqSongInfo `json:"list"`
		} `json:"song"`
	} `json:"data`
}

type qqSongInfo struct {
	SongID   int    `json:"songid"`
	SongMID  string `json:"songmid"`
	SongName string `json:"songname"`
	Singer   []struct {
		Name string `json:"name"`
	} `json:"singer"`
}

// 提取JSON，为QQ音乐准备
func extractJSONFromJSONP(jsonp string) string {
	start := strings.Index(jsonp, "(")
	end := strings.LastIndex(jsonp, ")")
	if start >= 0 && end > start {
		return jsonp[start+1 : end]
	}
	return ""
}

func QQConnectSinger(songList *qqSongInfo) string {
	ret := ""
	for _, singer := range songList.Singer {
		ret += singer.Name + ", "
	}
	return ret[:len(ret)-2]
}

// 给标题，返回id, mid, title, artist的切片指针
func QQGetMusic(Title string) (*[]string, *[]string, *[]string, *[]string, error) {
	url := strings.ReplaceAll(fmt.Sprintf(qqSearchIDTemplate, url.PathEscape(Title)), ` `, "%20")
	req, _ := http.NewRequest("GET", url, nil)
	SetHeader(req)
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		fmt.Println(err)
		return nil, nil, nil, nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		fmt.Println(resp.Status)
		return nil, nil, nil, nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var response qqSerachIDResponse
	_ = json.Unmarshal(body, &response)
	songList := response.Data.Song.List
	var ids, mids, names, artists []string
	for _, songInfo := range songList {
		if songInfo.SongID == 0 {
			continue
		}
		ids = append(ids, strconv.Itoa(songInfo.SongID))
		mids = append(mids, songInfo.SongMID)
		names = append(names, songInfo.SongName)
		artists = append(artists, QQConnectSinger(&songInfo))
	}
	return &ids, &mids, &names, &artists, nil
}

type qqLrcResponse struct {
	Data struct {
		Encode struct {
			Context string `json:"context"`
		} `json:"encode"`
	} `json:"data"`
}

// 通过ID查找歌词
// 返回歌词，翻译歌词以及错误
// 返回的歌词都是未转译的
func QQGetLyric(ID, MID string) (string, string, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf(qqLrcURLTemplate, ID, MID), nil)
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		fmt.Println(err)
		return "", "", errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		fmt.Println(resp.Status)
		return "", "", errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var response qqLrcResponse
	_ = json.Unmarshal(body, &response)
	return response.Data.Encode.Context, "", nil
}
