package file

import (
	"os"
	"sort"
	"strconv"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (info *File) WriteLyric() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB(db)

	return errors.WithStack(db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(lyricCacheKeyFilter(info.FolderName)).Delete(&lyricRow{}).Error; err != nil {
			return err
		}
		rows := make([]lyricRow, 0, len(info.InfoLyric))
		for _, infoLyric := range info.InfoLyric {
			lyricType := infoLyric.Type
			if lyricType == "" {
				lyricType = "lrc"
			}
			// 存储时统一按内容计算 is_complete，保证与响应/补全逻辑一致
			complete := computeComplete(infoLyric.Source, infoLyric.Lyric, infoLyric.Romaji)
			rows = append(rows, lyricRow{
				CacheKey:   info.FolderName,
				LyricID:    infoLyric.ID,
				Title:      infoLyric.Title,
				Artist:     infoLyric.Artist,
				Lyric:      infoLyric.Lyric,
				Romaji:     infoLyric.Romaji,
				Type:       lyricType,
				Source:     infoLyric.Source,
				IsComplete: boolToInt(complete),
			})
		}
		if len(rows) == 0 {
			return nil
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
		return nil
	}))
}

func (info *File) ReadLyric() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB(db)

	var rows []lyricRow
	if err := db.Where(lyricCacheKeyFilter(info.FolderName)).Find(&rows).Error; err != nil {
		return errors.WithStack(err)
	}
	if len(rows) == 0 {
		return errors.WithStack(os.ErrNotExist)
	}
	sort.Slice(rows, func(i, j int) bool {
		return lessLyricID(rows[i].LyricID, rows[j].LyricID)
	})
	for _, row := range rows {
		info.InfoLyric = append(info.InfoLyric, InfoLyric{
			ID:         row.LyricID,
			Title:      row.Title,
			Artist:     row.Artist,
			Lyric:      row.Lyric,
			Romaji:     row.Romaji,
			Type:       row.Type,
			Source:     row.Source,
			IsComplete: row.IsComplete != 0,
		})
	}
	info.HasPrevious = true
	return nil
}

func (info *File) RemoveLyric() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB(db)
	return errors.WithStack(db.Where(lyricCacheKeyFilter(info.FolderName)).Delete(&lyricRow{}).Error)
}

func lessLyricID(left, right string) bool {
	leftInt, leftErr := strconv.Atoi(left)
	rightInt, rightErr := strconv.Atoi(right)
	if leftErr == nil && rightErr == nil && leftInt != rightInt {
		return leftInt < rightInt
	}
	return left < right
}
