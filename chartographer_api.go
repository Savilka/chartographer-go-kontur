package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type ChartographerService struct {
	Router *gin.Engine
	DB     *sql.DB
}

type Charta struct {
	Width  int `form:"width" binding:"required"`
	Height int `form:"height" binding:"required"`
	Id     int
}

func (cs *ChartographerService) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, cs.Router))
}

func (cs *ChartographerService) Initialize() {
	var err error
	cs.DB, err = sql.Open("sqlite3", "chartas")
	if err != nil {
		log.Fatal(err)
	}

	cs.Router = gin.Default()

	cs.initEndpoints()

}

func (cs *ChartographerService) initEndpoints() {
	cs.Router.POST("/chartas/", cs.createChartaEndpoint)
	cs.Router.GET("/chartas/:id/", cs.addFragmentEndpoint)
	cs.Router.POST("/chartas/:id/", cs.getFragmentEndpoint)
	cs.Router.DELETE("/chartas/:id/", cs.deleteChartaEndpoint)
}

func (cs *ChartographerService) createChartaEndpoint(c *gin.Context) {
	var newCharta Charta
	if err := c.BindQuery(&newCharta); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, newCharta)
}

func (cs *ChartographerService) addFragmentEndpoint(c *gin.Context) {
}

func (cs *ChartographerService) getFragmentEndpoint(c *gin.Context) {
}

func (cs *ChartographerService) deleteChartaEndpoint(c *gin.Context) {

}
