package file

import (
	"database/sql"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
	"os"
)

// 文件包负责读写
type File struct {
	HasPrevious bool
	FolderName  string
	InfoLyric   []InfoLyric
}

type InfoLyric struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Lyric  string `json:"lyrics"`
}

func init() {
	os.Mkdir("assets", 0777)
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "assets/lyrics.db")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, errors.WithStack(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS lyrics (
cache_key TEXT NOT NULL,
lyric_id TEXT NOT NULL,
title TEXT NOT NULL,
artist TEXT NOT NULL,
lyric TEXT NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (cache_key, lyric_id)
);
CREATE INDEX IF NOT EXISTS idx_lyrics_cache_key ON lyrics(cache_key);`); err != nil {
		_ = db.Close()
		return nil, errors.WithStack(err)
	}
	return db, nil
}
