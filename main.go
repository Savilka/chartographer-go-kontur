package main

import (
	"github.com/gin-gonic/gin"
)

type Charta struct {
	Width  int `form:"width" binding:"required"`
	Height int `form:"height" binding:"required"`
	Id     int
}

func createChartaEndpoint(ctx *gin.Context) {

}

func addFragmentEndpoint(ctx *gin.Context) {

}

func getFragmentEndpoint(ctx *gin.Context) {

}

func deleteChartaEndpoint(ctx *gin.Context) {

}

func main() {
	router := gin.Default()

	router.POST("/chartas/", createChartaEndpoint)
	router.GET("/chartas/:id/", addFragmentEndpoint)
	router.POST("/chartas/:id/", getFragmentEndpoint)
	router.DELETE("/chartas/:id/", deleteChartaEndpoint)
}
