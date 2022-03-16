package main

import (
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/image/bmp"
	"image"
	"image/draw"
	"log"
	"net/http"
	"os"
	"strconv"
)

type ChartographerService struct {
	Router *gin.Engine
	DB     *bolt.DB
}

type Charta struct {
	Width  int `form:"width" binding:"required,gte=1,lte=20000"`
	Height int `form:"height" binding:"required,gte=1,lte=50000"`
	Id     string
}

type Fragment struct {
	Width  int  `form:"width" binding:"required,gte=1,lte=5000"`
	Height int  `form:"height" binding:"required,gte=1,lte=5000"`
	X      *int `form:"x" binding:"required"`
	Y      *int `form:"y" binding:"required"`
}

func (cs *ChartographerService) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, cs.Router))
}

func (cs *ChartographerService) Initialize() {
	var err error
	cs.DB, err = bolt.Open("chartas.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = cs.DB.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("chartas"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	cs.Router = gin.Default()

	cs.initEndpoints()

}

func (cs *ChartographerService) initEndpoints() {
	cs.Router.POST("/chartas/", cs.createChartaEndpoint)
	cs.Router.POST("/chartas/:id/", cs.addFragmentEndpoint)
	cs.Router.GET("/chartas/:id/", cs.getFragmentEndpoint)
	cs.Router.DELETE("/chartas/:id/", cs.deleteChartaEndpoint)
}

func (cs *ChartographerService) createChartaEndpoint(c *gin.Context) {
	var newCharta Charta
	if err := c.BindQuery(&newCharta); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := cs.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("chartas"))

		id, _ := b.NextSequence()
		newCharta.Id = strconv.Itoa(int(id))

		chartaImg := createBlackImage(newCharta.Width, newCharta.Height)
		filename := fmt.Sprintf("chartas/%s.bmp", newCharta.Id)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		err = Encode(file, chartaImg)
		if err != nil {
			return err
		}
		_ = file.Close()

		buf, err := json.Marshal(newCharta)
		if err != nil {
			return err
		}

		return b.Put([]byte(newCharta.Id), buf)
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (cs *ChartographerService) addFragmentEndpoint(c *gin.Context) {
}

func (cs *ChartographerService) getFragmentEndpoint(c *gin.Context) {
	var fragment Fragment
	if err := c.BindQuery(&fragment); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	x, y := *fragment.X, *fragment.Y

	err := cs.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("chartas"))

		v := b.Get([]byte(c.Param("id")))
		if v == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return nil
		}

		var charta Charta
		err := json.Unmarshal(v, &charta)
		if err != nil {
			return err
		}

		filename := fmt.Sprintf("chartas/%s.bmp", charta.Id)
		chartaImgBmp, err := os.Open(filename)
		if err != nil {
			return err
		}
		chartaImg, err := bmp.Decode(chartaImgBmp)
		if err != nil {
			return err
		}

		var fragmentOfChartaImg *image.NRGBA
		var fragmentImgBg *image.NRGBA

		switch {
		case x < 0 && y < 0:
			if fragment.Width <= -x || fragment.Width <= -y {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			fragmentImgBg = createBlackImage(fragment.Width, fragment.Height)
			//14
			if fragment.Width-x >= charta.Width && fragment.Height-y >= charta.Height {
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: -x, Y: -y},
					Max: image.Point{X: -x + fragment.Width, Y: -y + fragment.Height},
				}, chartaImg, image.Point{}, draw.Src)
			}
			//(15 and 10) and 16
			if fragment.Height+y <= charta.Height {
				if fragment.Width-x > charta.Width {
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, 0, charta.Width, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: -y},
						Max: image.Point{X: -x + fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Src)
				} else {
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, 0, charta.Width+x, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: -y},
						Max: image.Point{X: -x + fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Src)
				}
			} else {
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, 0, fragment.Width+x, fragment.Height))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: -x, Y: -y},
					Max: image.Point{X: fragment.Width, Y: charta.Height - y},
				}, fragmentOfChartaImg, image.Point{}, draw.Src)
			}
			err := writeBmpIntoTheBody(fragmentImgBg, c)
			if err != nil {
				return err
			}

		case x >= 0 && y < 0:
			if fragment.Height <= -y || charta.Width <= x {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}
			fragmentImgBg = createBlackImage(fragment.Width, fragment.Height)
			//(11 and 6) and 18
			if fragment.Height+y <= charta.Height {
				if x+fragment.Width > charta.Width {
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, 0, charta.Width, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: 0, Y: -y},
						Max: image.Point{X: charta.Width + x, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Src)
				} else {
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, 0, fragment.Width+x, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: 0, Y: -y},
						Max: image.Point{X: fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Src)
				}
			} else {
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, 0, charta.Width, charta.Height))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: 0, Y: -y},
					Max: image.Point{X: charta.Width - x, Y: charta.Height - y},
				}, fragmentOfChartaImg, image.Point{}, draw.Src)
			}

			err := writeBmpIntoTheBody(fragmentImgBg, c)
			if err != nil {
				return err
			}

		case x < 0 && y >= 0:
			if fragment.Width <= -x || charta.Height <= y {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			fragmentImgBg = createBlackImage(fragment.Width, fragment.Height)
			//(12 and 7) and 17
			if fragment.Width+x <= charta.Width {
				if fragment.Height+y >= charta.Height {
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, y, fragment.Width+x, charta.Height))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: 0},
						Max: image.Point{X: fragment.Width, Y: charta.Height - y},
					}, fragmentOfChartaImg, image.Point{}, draw.Src)
				} else {
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, y, fragment.Width+x, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: 0},
						Max: image.Point{X: fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Src)
				}
			} else {
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, y, charta.Width, charta.Height))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: -x, Y: 0},
					Max: image.Point{X: charta.Width - x, Y: charta.Height - y},
				}, fragmentOfChartaImg, image.Point{}, draw.Src)
			}

			err := writeBmpIntoTheBody(fragmentImgBg, c)
			if err != nil {
				return err
			}

		case x >= 0 && y >= 0:
			if x >= charta.Width || y >= charta.Height {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			fragmentImgBg = createBlackImage(fragment.Width, fragment.Height)

			if fragment.Width+x <= charta.Width && fragment.Height+y <= charta.Height {
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, fragment.Width+x, fragment.Height+y))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: fragment.Width, Y: fragment.Height},
				}, fragmentOfChartaImg, image.Point{}, draw.Src)
			} else {
				if fragment.Width+x >= charta.Width && fragment.Height+y >= charta.Height {
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, charta.Width, charta.Height))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{X: charta.Width - x, Y: charta.Height - y},
					}, fragmentOfChartaImg, image.Point{}, draw.Src)
				} else {
					if fragment.Width+x >= charta.Width {
						fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, fragment.Width, fragment.Height+y))
						draw.Draw(fragmentImgBg, image.Rectangle{
							Min: image.Point{X: 0, Y: 0},
							Max: image.Point{X: charta.Width - x, Y: fragment.Height},
						}, fragmentOfChartaImg, image.Point{}, draw.Src)
					} else {
						fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, fragment.Width+x, fragment.Height))
						draw.Draw(fragmentImgBg, image.Rectangle{
							Min: image.Point{X: 0, Y: 0},
							Max: image.Point{X: fragment.Width, Y: charta.Height - x},
						}, fragmentOfChartaImg, image.Point{}, draw.Src)
					}
				}
			}

			err := writeBmpIntoTheBody(fragmentImgBg, c)
			if err != nil {
				return err
			}

		}

		return nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

}

func (cs *ChartographerService) deleteChartaEndpoint(c *gin.Context) {
	id := c.Param("id")

	err := cs.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("chartas"))
		filename := fmt.Sprintf("chartas/%s.bmp", id)
		err := os.Remove(filename)
		if err != nil {
			return err
		}
		return b.Delete([]byte(id))
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.AbortWithStatus(200)
}
