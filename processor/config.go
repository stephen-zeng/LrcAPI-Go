package processor

var (
	API_TOT           = 3
	NETEASE_API_COUNT = 0
	QQ_API_COUNT      = 1
	KUGOU_API_COUNT   = 2
)

type Processor struct {
	InfoLyric []InfoLyric
	Title     string `json:"title"`
	Artist    string `json:"artist"`
}

type InfoLyric struct {
	Index  int    `json:"index"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Lyric  string `json:"lyrics"`
	Romaji string `json:"romaji"`
	Type   string `json:"type"`
	Source string `json:"source"`
	API    string `json:"api"`
}
