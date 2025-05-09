package util

import (
	"regexp"
	"strings"
)

func makeLrcMap(lrc string, timeStampPos [][]int, timeStamps []string) *map[string]string {
	ret := make(map[string]string)
	for index, timeStamp := range timeStamps {
		if index == len(timeStamps)-1 {
			ret[timeStamp] = lrc[timeStampPos[index][1]:]
		} else {
			ret[timeStamp] = lrc[timeStampPos[index][1]:timeStampPos[index+1][0]]
		}
	}
	return &ret
}

// 需要未转译的数据，也就是直接从netease获取的数据
func LrcTranslationBlender(LRC, TLRC string) string {
	re := regexp.MustCompile(`\[\d{2}:\d{2}\.\d{3}\]`)
	LRCTimeStamps := re.FindAllString(LRC, -1)
	if len(LRCTimeStamps) == 0 {
		re = regexp.MustCompile(`\[\d{2}:\d{2}\.\d{2}\]`)
		LRCTimeStamps = re.FindAllString(LRC, -1)
	}
	ret := ""
	TLRCMap := makeLrcMap(TLRC, re.FindAllStringIndex(TLRC, -1), re.FindAllString(TLRC, -1))
	LRCMap := makeLrcMap(LRC, re.FindAllStringIndex(LRC, -1), re.FindAllString(LRC, -1))
	for _, timeStamp := range LRCTimeStamps {
		ret += timeStamp + (*LRCMap)[timeStamp]
		if tlrc, exist := (*TLRCMap)[timeStamp]; exist {
			ret += timeStamp + tlrc
		}
	}
	return strings.ReplaceAll(ret, `\"`, `"`)
}
