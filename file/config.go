package file

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// 文件包负责读写
type File struct {
	HasPrevious bool
	FolderName  string
	InfoLyric   []InfoLyric
}

type InfoLyric struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	Lyric      string `json:"lyrics"`
	Romaji     string `json:"romaji"`
	Type       string `json:"type"`
	Source     string `json:"source"`
	IsComplete bool   `json:"isComplete"`
}

var lyricsDBPath = filepath.Join("assets", "lyrics.db")

func openDB() (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(lyricsDBPath), 0777); err != nil {
		return nil, errors.WithStack(err)
	}
	// busy_timeout 避免后台补全与在线读写并发时的 "database is locked"；
	// WAL 让读写可以并发进行。
	db, err := gorm.Open(sqlite.Open(lyricsDBPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"), &gorm.Config{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, errors.WithStack(err)
	}
	if err := migrateLyricsDB(db); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
