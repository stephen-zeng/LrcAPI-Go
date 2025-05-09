package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var (
	pwd string
	r   *gin.Engine
)

func Handler(port, passwd string) {
	pwd = passwd
	gin.SetMode(gin.ReleaseMode)
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST"}
	r = gin.New()
	r.Use(cors.New(config))
	r.GET("/lrc", lyricHandler)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{})
	})
	log.Printf("The server will be running on port %s", port)
	log.Printf("The default password is 123456")
	r.Run(":" + port)
}
