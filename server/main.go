package main

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
)

func main() {
	logFile, err := os.OpenFile("server.log", os.O_CREATE | os.O_RDWR | os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)
	r := gin.Default()
	r.Static("/", "public")

	log.Fatal(autotls.Run(r, "www.oshirisu.site"))
}