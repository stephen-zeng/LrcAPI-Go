package processor

import (
	"github.com/pkg/errors"
	"lrcAPI/util"
	"strconv"
)

func (data *Processor) Process() error {
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
		infoLyric := InfoLyric{
			ID:     strconv.Itoa(index),
			Title:  (*titles)[index],
			Artist: (*artists)[index],
			Lyric:  util.LrcTranslationBlender(oriLrc, trLrc),
		}
		data.InfoLyric = append(data.InfoLyric, infoLyric)
	}
	return nil
}
