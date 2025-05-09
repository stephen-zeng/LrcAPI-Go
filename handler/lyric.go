package handler

import (
	"github.com/gin-gonic/gin"
	"log"
	"lrcAPI/file"
	"lrcAPI/request"
	"lrcAPI/util"
	"net/http"
)

func lyricHandler(c *gin.Context) {
	if c.Request.Header.Get("Authorization") != pwd {
		log.Println("authorization required")
		c.JSON(404, gin.H{})
		return
	}
	var lyricRequest request.Request
	lyricRequest.Processor.Title = c.Query("title")
	lyricRequest.Processor.Artist = c.Query("artist")
	delOp := c.Query("delOp")
	lyricRequest.File.FolderName = lyricRequest.Processor.Artist + " - " + lyricRequest.Processor.Title
	if delOp == "true" {
		log.Printf("Delete Tmp for %s\n", lyricRequest.Processor.Artist)
		c.JSON(http.StatusOK, gin.H{
			"Delete": lyricRequest.File.FolderName,
		})
	}
	if err := lyricRequest.File.ReadLyric(); err == nil {
		log.Println("found exist")
		c.JSON(http.StatusOK, lyricRequest.File.InfoLyric)
		return
	}
	if err := lyricRequest.Processor.Process(); err != nil {
		util.ErrorPrinter(err)
		c.JSON(404, gin.H{})
	}
	for _, value := range lyricRequest.Processor.InfoLyric {
		lyricRequest.File.InfoLyric = append(lyricRequest.File.InfoLyric, file.InfoLyric{
			ID:     value.ID,
			Title:  value.Title,
			Artist: value.Artist,
			Lyric:  value.Lyric,
		})
	}
	if err := lyricRequest.File.WriteLyric(); err != nil {
		util.ErrorPrinter(err)
	}
	c.JSON(http.StatusOK, lyricRequest.File.InfoLyric)
}
