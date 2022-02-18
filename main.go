package main

import (
	"github.com/gin-gonic/gin"
)

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

	router.POST("/chartas/?width=:width&height=:height", createChartaEndpoint)
	router.GET("/chartas/:id/?x=:x&y=:y&width=:width&height=:height", addFragmentEndpoint)
	router.POST("/chartas/:id/?x=:x&y=:y&width=:width&height=:height", getFragmentEndpoint)
	router.DELETE("/chartas/:id/?x=:x&y=:y&width=:width&height=:height", deleteChartaEndpoint)
}
