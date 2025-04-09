package processor

type Processor struct {
	InfoLyric []InfoLyric
	Title     string `json:"title"`
	Artist    string `json:"artist"`
}

type InfoLyric struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Lyric  string `json:"lyrics"`
}
