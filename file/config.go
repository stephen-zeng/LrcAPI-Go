package file

import "os"

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
