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
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
)

type ChartographerService struct {
	Router   *gin.Engine
	DB       *bolt.DB
	pathName string
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

func (cs *ChartographerService) Initialize(path, dbName string) {
	var err error
	cs.pathName = path
	db := fmt.Sprintf("%s/%s", path, dbName)
	cs.DB, err = bolt.Open(db, 0600, nil)
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
		filename := fmt.Sprintf("%s/chartas/%s.png", cs.pathName, newCharta.Id)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		enc := &png.Encoder{
			CompressionLevel: png.NoCompression,
		}

		err = enc.Encode(file, chartaImg)
		if err != nil {
			return err
		}
		_ = file.Close()

		buf, err := json.Marshal(newCharta)
		if err != nil {
			return err
		}

		c.String(http.StatusCreated, newCharta.Id)
		return b.Put([]byte(newCharta.Id), buf)
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

}

func (cs *ChartographerService) addFragmentEndpoint(c *gin.Context) {
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

		filename := fmt.Sprintf("%s/chartas/%s.png", cs.pathName, charta.Id)
		chartaImgPng, err := os.OpenFile(filename, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		chartaImgRaw, err := png.Decode(chartaImgPng)
		if err != nil {
			return err
		}
		_ = chartaImgPng.Close()

		fragmentImg, err := bmp.Decode(c.Request.Body)
		if err != nil {
			return err
		}
		chartaImg := image.NewNRGBA(image.Rect(0, 0, charta.Width, charta.Height))
		draw.Draw(chartaImg, image.Rect(0, 0, charta.Width, charta.Height), chartaImgRaw, image.Point{}, draw.Over)

		var fragmentOfFragmentImg *image.NRGBA

		switch {
		case x < 0 && y < 0:
			//case 1
			if fragment.Width <= -x || fragment.Width <= -y {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			//case 24
			if fragment.Width-x >= charta.Width && fragment.Height-y >= charta.Height {
				fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(-x, -y, charta.Width-x, charta.Height-x))
				draw.Draw(chartaImg, image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: charta.Width, Y: charta.Height},
				}, fragmentOfFragmentImg, image.Point{}, draw.Over)
				break
			}

			if fragment.Height+y <= charta.Height {
				if fragment.Width-x > charta.Width {
					//case 11
					fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(-x, -y, charta.Width-x, fragment.Height))
					draw.Draw(chartaImg, image.Rectangle{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{X: charta.Width, Y: fragment.Height + y},
					}, fragmentOfFragmentImg, image.Point{}, draw.Over)
					break
				} else {
					//case 9
					fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(-x, -y, fragment.Width, fragment.Height))
					draw.Draw(chartaImg, image.Rectangle{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{X: fragment.Width + x, Y: fragment.Height + y},
					}, fragmentOfFragmentImg, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 10
				fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(-x, -y, fragment.Width, fragment.Height-y))
				draw.Draw(chartaImg, image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: fragment.Width + x, Y: fragment.Height},
				}, fragmentOfFragmentImg, image.Point{}, draw.Over)
				break
			}

		case x >= 0 && y < 0:
			//case 2 and 3
			if fragment.Height <= -y || charta.Width <= x {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			if fragment.Height+y <= charta.Height {
				if x+fragment.Width > charta.Width {
					//case 13
					fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(0, -y, charta.Width-x, fragment.Height))
					draw.Draw(chartaImg, image.Rectangle{
						Min: image.Point{X: x, Y: 0},
						Max: image.Point{X: charta.Width, Y: fragment.Height + y},
					}, fragmentOfFragmentImg, image.Point{}, draw.Over)
					break
				} else {
					//case 12
					fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(0, -y, fragment.Width, fragment.Height))
					draw.Draw(chartaImg, image.Rectangle{
						Min: image.Point{X: x, Y: 0},
						Max: image.Point{X: fragment.Width + x, Y: fragment.Height + y},
					}, fragmentOfFragmentImg, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 14 and 15
				fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(0, -y, charta.Width, charta.Height))
				draw.Draw(chartaImg, image.Rectangle{
					Min: image.Point{X: x, Y: 0},
					Max: image.Point{X: charta.Width, Y: charta.Height},
				}, fragmentOfFragmentImg, image.Point{}, draw.Over)
				break
			}

		case x < 0 && y >= 0:
			//case 4 and 5
			if fragment.Width <= -x || charta.Height <= y {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			if fragment.Width+x <= charta.Width {
				if fragment.Height+y >= charta.Height {
					//case 17
					fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(-x, 0, fragment.Width, charta.Height-y))
					draw.Draw(chartaImg, image.Rectangle{
						Min: image.Point{X: 0, Y: y},
						Max: image.Point{X: fragment.Width + x, Y: charta.Height},
					}, fragmentOfFragmentImg, image.Point{}, draw.Over)
					break
				} else {
					//case 16
					fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(-x, 0, fragment.Width, fragment.Height-y))
					draw.Draw(chartaImg, image.Rectangle{
						Min: image.Point{X: 0, Y: y},
						Max: image.Point{X: fragment.Width + x, Y: fragment.Height + y},
					}, fragmentOfFragmentImg, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 18 and 19
				fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(-x, 0, charta.Width-x, charta.Height-y))
				draw.Draw(chartaImg, image.Rectangle{
					Min: image.Point{X: 0, Y: y},
					Max: image.Point{X: charta.Width, Y: charta.Height},
				}, fragmentOfFragmentImg, image.Point{}, draw.Over)
				break
			}

		case x >= 0 && y >= 0:
			//case 6, 7 and 8
			if x >= charta.Width || y >= charta.Height {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			//case 20
			if fragment.Width+x <= charta.Width && fragment.Height+y <= charta.Height {
				fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(0, 0, fragment.Width, fragment.Height))
				draw.Draw(chartaImg, image.Rectangle{
					Min: image.Point{X: x, Y: y},
					Max: image.Point{X: fragment.Width + x, Y: fragment.Height + y},
				}, fragmentOfFragmentImg, image.Point{}, draw.Over)
				break
			} else {
				if fragment.Width+x >= charta.Width && fragment.Height+y >= charta.Height {
					//case 23
					fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(0, 0, charta.Width-x, charta.Height-y))
					draw.Draw(chartaImg, image.Rectangle{
						Min: image.Point{X: x, Y: y},
						Max: image.Point{X: charta.Width, Y: charta.Height},
					}, fragmentOfFragmentImg, image.Point{}, draw.Over)
					break
				} else {
					if fragment.Width+x >= charta.Width {
						//case 21
						fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(0, 0, charta.Width-x, fragment.Height))
						draw.Draw(chartaImg, image.Rectangle{
							Min: image.Point{X: x, Y: y},
							Max: image.Point{X: charta.Width, Y: fragment.Height + y},
						}, fragmentOfFragmentImg, image.Point{}, draw.Over)
						break
					} else {
						//case 22
						fragmentOfFragmentImg = imaging.Crop(fragmentImg, image.Rect(0, 0, fragment.Width, charta.Height-y))
						draw.Draw(chartaImg, image.Rectangle{
							Min: image.Point{X: x, Y: y},
							Max: image.Point{X: fragment.Width + y, Y: charta.Height},
						}, fragmentOfFragmentImg, image.Point{}, draw.Over)
						break
					}
				}
			}
		}

		enc := &png.Encoder{
			CompressionLevel: png.NoCompression,
		}
		chartaImgPng, err = os.OpenFile(filename, os.O_RDWR, 0644)
		err = enc.Encode(chartaImgPng, chartaImg)
		if err != nil {
			return err
		}
		_ = chartaImgPng.Close()

		return nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
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

		filename := fmt.Sprintf("%s/chartas/%s.png", cs.pathName, charta.Id)
		chartaImgPng, err := os.Open(filename)
		if err != nil {
			return err
		}
		chartaImg, err := png.Decode(chartaImgPng)
		if err != nil {
			return err
		}

		var fragmentOfChartaImg *image.NRGBA
		var fragmentImgBg *image.NRGBA
		fragmentImgBg = createBlackImage(fragment.Width, fragment.Height)

		switch {
		case x < 0 && y < 0:
			//case 1
			if fragment.Width <= -x || fragment.Width <= -y {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			//case 24
			if fragment.Width-x >= charta.Width && fragment.Height-y >= charta.Height {
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: -x, Y: -y},
					Max: image.Point{X: -x + charta.Width, Y: -y + charta.Height},
				}, chartaImg, image.Point{}, draw.Over)
				break
			}

			if fragment.Height+y <= charta.Height {
				if fragment.Width-x > charta.Width {
					//case 11
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, 0, charta.Width, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: -y},
						Max: image.Point{X: -x + fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Over)
					break
				} else {
					//case 9
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, 0, charta.Width+x, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: -y},
						Max: image.Point{X: -x + fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 10
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, 0, fragment.Width+x, fragment.Height))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: -x, Y: -y},
					Max: image.Point{X: fragment.Width, Y: charta.Height - y},
				}, fragmentOfChartaImg, image.Point{}, draw.Over)
				break
			}

		case x >= 0 && y < 0:
			//case 2 and 3
			if fragment.Height <= -y || charta.Width <= x {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			if fragment.Height+y <= charta.Height {
				if x+fragment.Width > charta.Width {
					//case 13
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, 0, charta.Width, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: 0, Y: -y},
						Max: image.Point{X: charta.Width + x, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Over)
					break
				} else {
					//case 12
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, 0, fragment.Width+x, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: 0, Y: -y},
						Max: image.Point{X: fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 14 and 15
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, 0, charta.Width, charta.Height))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: 0, Y: -y},
					Max: image.Point{X: charta.Width - x, Y: charta.Height - y},
				}, fragmentOfChartaImg, image.Point{}, draw.Over)
				break
			}

		case x < 0 && y >= 0:
			//case 4 and 5
			if fragment.Width <= -x || charta.Height <= y {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			if fragment.Width+x <= charta.Width {
				if fragment.Height+y >= charta.Height {
					//case 17
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, y, fragment.Width+x, charta.Height))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: 0},
						Max: image.Point{X: fragment.Width, Y: charta.Height - y},
					}, fragmentOfChartaImg, image.Point{}, draw.Over)
					break
				} else {
					//case 16
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, y, fragment.Width+x, fragment.Height+y))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: -x, Y: 0},
						Max: image.Point{X: fragment.Width, Y: fragment.Height},
					}, fragmentOfChartaImg, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 18 and 19
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(0, y, charta.Width, charta.Height))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: -x, Y: 0},
					Max: image.Point{X: charta.Width - x, Y: charta.Height - y},
				}, fragmentOfChartaImg, image.Point{}, draw.Over)
				break
			}

		case x >= 0 && y >= 0:
			//case 6, 7 and 8
			if x >= charta.Width || y >= charta.Height {
				c.AbortWithStatus(http.StatusBadRequest)
				return nil
			}

			//case 20
			if fragment.Width+x <= charta.Width && fragment.Height+y <= charta.Height {
				fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, fragment.Width+x, fragment.Height+y))
				draw.Draw(fragmentImgBg, image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: fragment.Width, Y: fragment.Height},
				}, fragmentOfChartaImg, image.Point{}, draw.Over)
				break
			} else {
				if fragment.Width+x >= charta.Width && fragment.Height+y >= charta.Height {
					//case 23
					fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, charta.Width, charta.Height))
					draw.Draw(fragmentImgBg, image.Rectangle{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{X: charta.Width - x, Y: charta.Height - y},
					}, fragmentOfChartaImg, image.Point{}, draw.Over)
					break
				} else {
					if fragment.Width+x >= charta.Width {
						//case 21
						fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, fragment.Width, fragment.Height+y))
						draw.Draw(fragmentImgBg, image.Rectangle{
							Min: image.Point{X: 0, Y: 0},
							Max: image.Point{X: charta.Width - x, Y: fragment.Height},
						}, fragmentOfChartaImg, image.Point{}, draw.Over)
						break
					} else {
						//case 22
						fragmentOfChartaImg = imaging.Crop(chartaImg, image.Rect(x, y, fragment.Width+x, fragment.Height))
						draw.Draw(fragmentImgBg, image.Rectangle{
							Min: image.Point{X: 0, Y: 0},
							Max: image.Point{X: fragment.Width, Y: charta.Height - x},
						}, fragmentOfChartaImg, image.Point{}, draw.Over)
						break
					}
				}
			}

		}

		err = writeBmpIntoTheBody(fragmentImgBg, c)
		if err != nil {
			return err
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
		filename := fmt.Sprintf("%s/chartas/%s.png", cs.pathName, id)
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

	c.Status(http.StatusOK)
}
