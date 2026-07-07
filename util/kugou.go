package util

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	kugouSearchTemplate  string = `https://mobileservice.kugou.com/api/v3/search/song?format=json&keyword=%s&page=1&pagesize=20&showtype=1`
	kugouKRCSearchTmpl   string = `https://krcs.kugou.com/search?ver=1&man=yes&client=mobi&hash=%s&album_audio_id=&duration=&lrctype=&keyword=`
	kugouKRCDownloadTmpl string = `https://lyrics.kugou.com/download?ver=1&client=pc&id=%s&accesskey=%s&fmt=krc&charset=utf8`
)

type kugouSearchResponse struct {
	Data struct {
		Info []struct {
			Hash       string `json:"hash"`
			SongName   string `json:"songname"`
			SingerName string `json:"singername"`
		} `json:"info"`
	} `json:"data"`
}

// KugouGetMusic 按关键词搜索，返回 hash / 曲名 / 歌手 切片。
func KugouGetMusic(keyword string) (*[]string, *[]string, *[]string, error) {
	u := strings.ReplaceAll(fmt.Sprintf(kugouSearchTemplate, url.QueryEscape(keyword)), `+`, "%20")
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "*/*")
	if KugouCookie != "" {
		req.Header.Set("Cookie", KugouCookie)
	}
	resp, err := GlobalHTTPClient.Do(req)
	switch {
	case err != nil:
		return nil, nil, nil, errors.WithStack(err)
	case resp.StatusCode != http.StatusOK:
		return nil, nil, nil, errors.WithStack(errors.New(resp.Status))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var response kugouSearchResponse
	_ = json.Unmarshal(body, &response)
	var hashes, names, artists []string
	for _, info := range response.Data.Info {
		if info.Hash == "" {
			continue
		}
		hashes = append(hashes, info.Hash)
		names = append(names, info.SongName)
		artists = append(artists, info.SingerName)
	}
	return &hashes, &names, &artists, nil
}

type kugouKRCSearchResponse struct {
	Status     int    `json:"status"`
	Info       string `json:"info"`
	Candidates []struct {
		ID        string `json:"id"`
		AccessKey string `json:"accesskey"`
	} `json:"candidates"`
}

type kugouKRCDownloadResponse struct {
	Status  int    `json:"status"`
	Content string `json:"content"`
	Fmt     string `json:"fmt"`
}

// KugouGetLyric 通过 hash 获取 KRC 逐词歌词并解析为统一结构。
func KugouGetLyric(hash string) (*LyricResult, error) {
	// 1. KRC 搜索拿到 id/accesskey
	req, _ := http.NewRequest("GET", fmt.Sprintf(kugouKRCSearchTmpl, hash), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "*/*")
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.WithStack(errors.New(resp.Status))
	}
	body, _ := io.ReadAll(resp.Body)
	var search kugouKRCSearchResponse
	if err := json.Unmarshal(body, &search); err != nil {
		return nil, errors.WithStack(err)
	}
	if len(search.Candidates) == 0 {
		return nil, errors.New("kugou: no krc candidate")
	}

	// 2. 下载 KRC
	cand := search.Candidates[0]
	dlURL := fmt.Sprintf(kugouKRCDownloadTmpl, cand.ID, cand.AccessKey)
	dlReq, _ := http.NewRequest("GET", dlURL, nil)
	dlReq.Header.Set("User-Agent", "Mozilla/5.0")
	dlReq.Header.Set("Accept", "*/*")
	if KugouCookie != "" {
		dlReq.Header.Set("Cookie", KugouCookie)
	}
	dlResp, err := GlobalHTTPClient.Do(dlReq)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer dlResp.Body.Close()
	if dlResp.StatusCode != http.StatusOK {
		return nil, errors.WithStack(errors.New(dlResp.Status))
	}
	dlBody, _ := io.ReadAll(dlResp.Body)
	var dl kugouKRCDownloadResponse
	if err := json.Unmarshal(dlBody, &dl); err != nil {
		return nil, errors.WithStack(err)
	}
	if strings.TrimSpace(dl.Content) == "" {
		return nil, errors.New("kugou: empty krc content")
	}

	// 3. 解密并解析
	krcText, err := DecodeKRC(dl.Content)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := ParseKRC(krcText)
	if strings.TrimSpace(result.Lyric) == "" {
		return nil, errors.New("kugou: empty parsed lyric")
	}
	return result, nil
}
