package file

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"sort"
)

func (info *File) WriteLyric() error {
	pathName := "assets/" + info.FolderName
	os.Mkdir(pathName, 0777)
	for _, infoLyric := range info.InfoLyric {
		lyricFile, err := os.Create(pathName + "/" + infoLyric.ID + ".json")
		if err != nil {
			return errors.WithStack(err)
		}
		defer lyricFile.Close()
		encoder := json.NewEncoder(lyricFile)
		encoder.Encode(infoLyric)
	}
	return nil
}

func (info *File) ReadLyric() error {
	pathName := "assets/" + info.FolderName
	lyricFolder, err := os.ReadDir(pathName)
	if len(lyricFolder) == 0 ||
		(len(lyricFolder) == 1 && lyricFolder[0].Name()[0] == '.') || // fuck you .DS_Store!!!
		(len(lyricFolder) == 1 && lyricFolder[0].Name() == "desktop.ini") { // fuck you desktop.ini!!!
		err = os.ErrNotExist
	}
	if err != nil {
		return errors.WithStack(err)
	}
	info.HasPrevious = true
	sort.Slice(lyricFolder, func(i, j int) bool {
		return lyricFolder[i].Name() < lyricFolder[j].Name()
	})
	for _, infoLyricFile := range lyricFolder {
		lyricFile, _ := os.Open(pathName + "/" + infoLyricFile.Name())
		defer lyricFile.Close()
		decoder := json.NewDecoder(lyricFile)
		var infoLyric InfoLyric
		decoder.Decode(&infoLyric)
		info.InfoLyric = append(info.InfoLyric, infoLyric)
	}
	return nil
}

func (info *File) RemoveLyric() error {
	pathName := "assets/" + info.FolderName
	return errors.WithStack(os.RemoveAll(pathName))
}
