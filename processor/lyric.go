package processor

import (
	"fmt"
	"github.com/pkg/errors"
	"lrcAPI/util"
	"sort"
)

func (data *Processor) Process() error {
	if err := netease(data); err != nil {
		return errors.Wrap(err, "neteaseAPI")
	}
	if err := qq(data); err != nil {
		return errors.Wrap(err, "qqAPI")
	}
	sort.Slice(data.InfoLyric, func(i, j int) bool {
		return data.InfoLyric[i].Index < data.InfoLyric[j].Index
	})
	data.InfoLyric = append(data.InfoLyric, InfoLyric{
		Index:  -1,
		Title:  data.Title,
		Artist: data.Artist,
		Lyric:  fmt.Sprintf("[00:00.00]%s\n[00:00.00]%s", data.Title, data.Artist),
	})
	return nil
}

func netease(data *Processor) error {
	musicIDs, titles, artists, err := util.NeteaseGetMusic(data.Title, data.Artist)
	if err != nil {
		return errors.WithStack(err)
	}
	for index, musicID := range *musicIDs {
		if index >= 5 {
			break
		}
		oriLrc, trLrc, err := util.NeteaseGetLyric(musicID)
		if err != nil {
			return errors.WithStack(err)
		}
		data.InfoLyric = append(data.InfoLyric, InfoLyric{
			Index:  API_TOT*index + NETEASE_API_COUNT,
			Title:  (*titles)[index] + " (网易云音乐)",
			Artist: (*artists)[index],
			Lyric:  util.LrcTranslationBlender(oriLrc, trLrc),
		})
	}
	return nil
}

func qq(data *Processor) error {
	musicIDs, musicMIDs, titles, artists, err := util.QQGetMusic(data.Title)
	if err != nil {
		return errors.WithStack(err)
	}
	for index, musicID := range *musicIDs {
		if index >= 5 {
			break
		}
		oriLrc, trLrc, err := util.QQGetLyric(musicID, (*musicMIDs)[index])
		if err != nil {
			return errors.WithStack(err)
		}
		data.InfoLyric = append(data.InfoLyric, InfoLyric{
			Index:  API_TOT*index + QQ_API_COUNT,
			Title:  (*titles)[index] + " (QQ音乐)",
			Artist: (*artists)[index],
			Lyric:  util.LrcTranslationBlender(oriLrc, trLrc),
		})
	}
	return nil
}
