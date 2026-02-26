package file

import (
	"github.com/pkg/errors"
	"os"
)

func (info *File) WriteLyric() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err := tx.Exec("DELETE FROM lyrics WHERE cache_key = ?", info.FolderName); err != nil {
		_ = tx.Rollback()
		return errors.WithStack(err)
	}
	stmt, err := tx.Prepare("INSERT INTO lyrics (cache_key, lyric_id, title, artist, lyric) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return errors.WithStack(err)
	}
	defer stmt.Close()
	for _, infoLyric := range info.InfoLyric {
		if _, err := stmt.Exec(info.FolderName, infoLyric.ID, infoLyric.Title, infoLyric.Artist, infoLyric.Lyric); err != nil {
			_ = tx.Rollback()
			return errors.WithStack(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (info *File) ReadLyric() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query("SELECT lyric_id, title, artist, lyric FROM lyrics WHERE cache_key = ? ORDER BY CAST(lyric_id AS INTEGER), lyric_id", info.FolderName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer rows.Close()
	for rows.Next() {
		var infoLyric InfoLyric
		if err := rows.Scan(&infoLyric.ID, &infoLyric.Title, &infoLyric.Artist, &infoLyric.Lyric); err != nil {
			return errors.WithStack(err)
		}
		info.InfoLyric = append(info.InfoLyric, infoLyric)
	}
	if err := rows.Err(); err != nil {
		return errors.WithStack(err)
	}
	if len(info.InfoLyric) == 0 {
		return errors.WithStack(os.ErrNotExist)
	}
	info.HasPrevious = true
	return nil
}

func (info *File) RemoveLyric() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM lyrics WHERE cache_key = ?", info.FolderName)
	return errors.WithStack(err)
}
