package processor

import (
	"github.com/pkg/errors"
	"lrcAPI/util"
	"strconv"
)

func (data *Processor) Process() error {
	var titles, artists, oriLrc, trLrc []string
	if err := netease(data, &titles, &artists, &oriLrc, &trLrc); err != nil {
		return errors.Wrap(err, "neteaseAPI")
	}
	if err := qq(data, &titles, &artists, &oriLrc, &trLrc); err != nil {
		return errors.Wrap(err, "qqAPI")
	}
	for index, _ := range titles {
		infoLyric := InfoLyric{
			ID:     strconv.Itoa(index),
			Title:  titles[index],
			Artist: artists[index],
			Lyric:  util.LrcTranslationBlender(oriLrc[index], trLrc[index]),
		}
		data.InfoLyric = append(data.InfoLyric, infoLyric)
	}
	return nil
}

func netease(data *Processor, Titles, Artists, OriLrc, TrLrc *[]string) error {
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
		*Titles = append(*Titles, (*titles)[index])
		*Artists = append(*Artists, (*artists)[index])
		*OriLrc = append(*OriLrc, oriLrc)
		*TrLrc = append(*TrLrc, trLrc)
	}
	return nil
}

func qq(data *Processor, Titles, Artists, OriLrc, TrLrc *[]string) error {
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
		*Titles = append(*Titles, (*titles)[index])
		*Artists = append(*Artists, (*artists)[index])
		*OriLrc = append(*OriLrc, oriLrc)
		*TrLrc = append(*TrLrc, trLrc)
	}
	return nil
}
