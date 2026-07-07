package processor

import (
	"fmt"
	"github.com/pkg/errors"
	"lrcAPI/util"
	"sort"
)

const maxPerSource = 5

func (data *Processor) Process() error {
	// 每个来源相互独立，单个来源失败不影响其它来源。
	if err := netease(data); err != nil {
		util.ErrorPrinter(errors.Wrap(err, "neteaseAPI"))
	}
	if err := qq(data); err != nil {
		util.ErrorPrinter(errors.Wrap(err, "qqAPI"))
	}
	if err := kugou(data); err != nil {
		util.ErrorPrinter(errors.Wrap(err, "kugouAPI"))
	}
	sort.Slice(data.InfoLyric, func(i, j int) bool {
		return data.InfoLyric[i].Index < data.InfoLyric[j].Index
	})
	data.InfoLyric = append(data.InfoLyric, InfoLyric{
		Index:  -1,
		Title:  data.Title,
		Artist: data.Artist,
		Lyric:  fmt.Sprintf("[00:00.00]%s\n[00:00.00]%s", data.Title, data.Artist),
		Type:   "lrc",
		Source: "fallback",
	})
	return nil
}

func netease(data *Processor) error {
	musicIDs, titles, artists, err := util.NeteaseGetMusic(data.Title, data.Artist)
	if err != nil {
		return errors.WithStack(err)
	}
	for index, musicID := range *musicIDs {
		if index >= maxPerSource {
			break
		}
		res, err := util.NeteaseGetLyric(musicID)
		if err != nil {
			return errors.WithStack(err)
		}
		data.InfoLyric = append(data.InfoLyric, InfoLyric{
			Index:  API_TOT*index + NETEASE_API_COUNT,
			Title:  (*titles)[index] + " (网易云音乐)",
			Artist: (*artists)[index],
			Lyric:  util.LrcTranslationBlender(res.Lyric, res.Translation),
			Romaji: res.Roma,
			Type:   res.Type,
			Source: "netease",
		})
	}
	return nil
}

func qq(data *Processor) error {
	musicIDs, musicMIDs, titles, artists, err := util.QQGetMusic(data.Title)
	if err != nil {
		return errors.WithStack(err)
	}
	for index := range *musicIDs {
		if index >= maxPerSource {
			break
		}
		res, err := util.QQGetLyricByMID((*musicMIDs)[index])
		if err != nil {
			// 单曲失败跳过，不中断整个来源
			continue
		}
		data.InfoLyric = append(data.InfoLyric, InfoLyric{
			Index:  API_TOT*index + QQ_API_COUNT,
			Title:  (*titles)[index] + " (QQ音乐)",
			Artist: (*artists)[index],
			Lyric:  util.LrcTranslationBlender(res.Lyric, res.Translation),
			Romaji: res.Roma,
			Type:   res.Type,
			Source: "qqmusic",
		})
	}
	return nil
}

func kugou(data *Processor) error {
	hashes, titles, artists, err := util.KugouGetMusic(data.Title + " " + data.Artist)
	if err != nil {
		return errors.WithStack(err)
	}
	for index, hash := range *hashes {
		if index >= maxPerSource {
			break
		}
		res, err := util.KugouGetLyric(hash)
		if err != nil {
			continue
		}
		data.InfoLyric = append(data.InfoLyric, InfoLyric{
			Index:  API_TOT*index + KUGOU_API_COUNT,
			Title:  (*titles)[index] + " (酷狗音乐)",
			Artist: (*artists)[index],
			Lyric:  util.LrcTranslationBlender(res.Lyric, res.Translation),
			Romaji: res.Roma,
			Type:   res.Type,
			Source: "kugou",
		})
	}
	return nil
}
