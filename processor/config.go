package processor

var (
	API_TOT           = 2
	NETEASE_API_COUNT = 0
	QQ_API_COUNT      = 1
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
	API    string `json:"api"`
}
