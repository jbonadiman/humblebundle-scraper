package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jbonadiman/webscrapers/api"
)

func main() {
	r := gin.Default()
	r.GET(
		"/api/scrap", func(c *gin.Context) {
			api.Handler(c.Writer, c.Request)
		},
	)
	err := r.Run(":3000")
	if err != nil {
		panic(err)
	}
}
