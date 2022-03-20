package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/image/bmp"
	"image"
	"image/draw"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type fragmentTest struct {
	Width  int
	Height int
	X      int
	Y      int
}

type chartaTest struct {
	Width  int
	Height int
}

var cs ChartographerService

func TestMain(m *testing.M) {
	cs.Initialize(".", "test.db")
	code := m.Run()
	err := cs.DB.Close()
	if err != nil {
		return
	}
	err = os.Remove("test.db")
	if err != nil {
		return
	}
	os.Exit(code)
}

func TestCreateChartaEndpoint(t *testing.T) {
	testCasesCreated := []chartaTest{
		{Width: 1, Height: 1},
		{Width: 1000, Height: 1000},
		{Width: 20000, Height: 50000},
	}

	testCasesBadRequest := []chartaTest{
		{Width: -100, Height: -100},
		{Width: -100, Height: 100},
		{Width: 100, Height: -100},
		{Width: 0, Height: 0},
		{Width: 25000, Height: 55000},
		{Width: 25000, Height: 20000},
		{Width: 10000, Height: 55000},
	}

	for _, testCase := range testCasesCreated {
		url := fmt.Sprintf("/chartas/?width=%d&height=%d", testCase.Width, testCase.Height)
		req, _ := http.NewRequest("POST", url, nil)
		response := httptest.NewRecorder()
		cs.Router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusCreated, response.Code)

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(response.Body)
		if err != nil {
			return
		}
		id := buf.String()

		url = fmt.Sprintf("/chartas/%s/", id)
		req, _ = http.NewRequest("DELETE", url, nil)
		responseDelete := httptest.NewRecorder()
		cs.Router.ServeHTTP(responseDelete, req)
		assert.Equal(t, http.StatusOK, responseDelete.Code)
	}

	for _, testCase := range testCasesBadRequest {
		url := fmt.Sprintf("/chartas/?width=%d&height=%d", testCase.Width, testCase.Height)
		req, _ := http.NewRequest("POST", url, nil)
		response := httptest.NewRecorder()
		cs.Router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}

	url := fmt.Sprintf("/chartas/?widthhhhhh=%d&heighttttttt=%d", 100, 100)
	req, _ := http.NewRequest("POST", url, nil)
	responseBadUrl := httptest.NewRecorder()
	cs.Router.ServeHTTP(responseBadUrl, req)
	assert.Equal(t, http.StatusBadRequest, responseBadUrl.Code)
}

func TestGetFragmentEndpoint(t *testing.T) {
	testCasesOK := []fragmentTest{
		{X: -25, Y: -25, Width: 50, Height: 50},   // 9
		{X: -25, Y: -25, Width: 50, Height: 150},  // 10
		{X: -25, Y: -25, Width: 150, Height: 50},  // 11
		{X: 25, Y: -25, Width: 50, Height: 50},    // 12
		{X: 75, Y: -25, Width: 50, Height: 50},    // 13
		{X: 25, Y: -25, Width: 150, Height: 150},  // 14
		{X: 25, Y: -25, Width: 50, Height: 150},   // 15
		{X: -25, Y: 25, Width: 50, Height: 50},    // 16
		{X: -25, Y: 75, Width: 50, Height: 50},    // 17
		{X: -25, Y: 75, Width: 150, Height: 50},   // 18
		{X: -25, Y: 25, Width: 150, Height: 50},   // 19
		{X: 25, Y: 25, Width: 50, Height: 50},     // 20
		{X: 25, Y: 25, Width: 150, Height: 50},    // 21
		{X: 25, Y: 25, Width: 50, Height: 150},    // 22
		{X: 25, Y: 25, Width: 150, Height: 150},   // 23
		{X: -25, Y: -25, Width: 150, Height: 150}, // 24
	}

	testCasesBadReq := []fragmentTest{
		{X: -100, Y: -100, Width: 50, Height: 50}, // 1
		{X: 40, Y: -60, Width: 50, Height: 50},    // 2
		{X: 150, Y: -100, Width: 50, Height: 50},  // 3
		{X: -100, Y: 50, Width: 50, Height: 50},   // 4
		{X: -100, Y: 150, Width: 50, Height: 50},  // 5
		{X: 150, Y: 40, Width: 50, Height: 50},    // 6
		{X: 150, Y: 150, Width: 50, Height: 50},   // 7
		{X: 40, Y: 150, Width: 50, Height: 50},    // 8
	}

	url := fmt.Sprintf("/chartas/?width=%d&height=%d", 100, 100)
	req, _ := http.NewRequest("POST", url, nil)
	responseCreate := httptest.NewRecorder()
	cs.Router.ServeHTTP(responseCreate, req)
	assert.Equal(t, http.StatusCreated, responseCreate.Code)

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(responseCreate.Body)
	if err != nil {
		return
	}
	id := buf.String()

	redFragment := createRedImage(100, 100)
	buf = new(bytes.Buffer)
	err = bmp.Encode(buf, redFragment)
	if err != nil {
		fmt.Println(err)
	}

	url = fmt.Sprintf("/chartas/%s/?x=%d&y=%d&width=%d&height=%d", id, 0, 0, 100, 100)
	req, _ = http.NewRequest("POST", url, buf)
	responseAdd := httptest.NewRecorder()
	cs.Router.ServeHTTP(responseAdd, req)
	assert.Equal(t, http.StatusOK, responseAdd.Code)

	filename := fmt.Sprintf("chartas/%s.png", id)
	fragmentRedFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}

	fragmentRed, err := png.Decode(fragmentRedFile)
	if err != nil {
		fmt.Println(err)
	}

	for _, testCase := range testCasesOK {

		fragmentBlack := createBlackImage(testCase.Width, testCase.Height)
		switch {
		case testCase.X < 0 && testCase.Y < 0:
			//case 24
			if testCase.Width-testCase.X >= 100 && testCase.Height-testCase.Y >= 100 {
				draw.Draw(&fragmentBlack, image.Rectangle{
					Min: image.Point{X: -testCase.X, Y: -testCase.Y},
					Max: image.Point{X: -testCase.X + 100, Y: -testCase.Y + 100},
				}, fragmentRed, image.Point{}, draw.Over)
				break
			}

			if testCase.Height+testCase.Y <= 100 {
				if testCase.Width-testCase.X > 100 {
					//case 11
					draw.Draw(&fragmentBlack, image.Rectangle{
						Min: image.Point{X: -testCase.X, Y: -testCase.Y},
						Max: image.Point{X: -testCase.X + testCase.Width, Y: testCase.Height},
					}, fragmentRed, image.Point{}, draw.Over)
					break
				} else {
					//case 9
					draw.Draw(&fragmentBlack, image.Rectangle{
						Min: image.Point{X: -testCase.X, Y: -testCase.Y},
						Max: image.Point{X: -testCase.X + testCase.Width, Y: testCase.Height},
					}, fragmentRed, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 10
				draw.Draw(&fragmentBlack, image.Rectangle{
					Min: image.Point{X: -testCase.X, Y: -testCase.Y},
					Max: image.Point{X: testCase.Width, Y: 100 - testCase.Y},
				}, fragmentRed, image.Point{}, draw.Over)
				break
			}

		case testCase.X >= 0 && testCase.Y < 0:

			if testCase.Height+testCase.Y <= 100 {
				if testCase.X+testCase.Width > 100 {
					//case 13
					draw.Draw(&fragmentBlack, image.Rectangle{
						Min: image.Point{X: 0, Y: -testCase.Y},
						Max: image.Point{X: 100 - testCase.X, Y: testCase.Height},
					}, fragmentRed, image.Point{}, draw.Over)
					break
				} else {
					//case 12
					draw.Draw(&fragmentBlack, image.Rectangle{
						Min: image.Point{X: 0, Y: -testCase.Y},
						Max: image.Point{X: testCase.Width, Y: testCase.Height},
					}, fragmentRed, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 14 and 15
				draw.Draw(&fragmentBlack, image.Rectangle{
					Min: image.Point{X: 0, Y: -testCase.Y},
					Max: image.Point{X: 100 - testCase.X, Y: 100 - testCase.Y},
				}, fragmentRed, image.Point{}, draw.Over)
				break
			}

		case testCase.X < 0 && testCase.Y >= 0:

			if testCase.Width+testCase.X <= 100 {
				if testCase.Height+testCase.Y >= 100 {
					//case 17
					draw.Draw(&fragmentBlack, image.Rectangle{
						Min: image.Point{X: -testCase.X, Y: 0},
						Max: image.Point{X: testCase.Width, Y: 100 - testCase.Y},
					}, fragmentRed, image.Point{}, draw.Over)
					break
				} else {
					//case 16
					draw.Draw(&fragmentBlack, image.Rectangle{
						Min: image.Point{X: -testCase.X, Y: 0},
						Max: image.Point{X: testCase.Width, Y: testCase.Height},
					}, fragmentRed, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 18 and 19
				draw.Draw(&fragmentBlack, image.Rectangle{
					Min: image.Point{X: -testCase.X, Y: 0},
					Max: image.Point{X: 100 - testCase.X, Y: 100 - testCase.Y},
				}, fragmentRed, image.Point{}, draw.Over)
				break
			}

		case testCase.X >= 0 && testCase.Y >= 0:

			//case 20
			if testCase.Width+testCase.X <= 100 && testCase.Height+testCase.Y <= 100 {
				draw.Draw(&fragmentBlack, image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: testCase.Width, Y: testCase.Height},
				}, fragmentRed, image.Point{}, draw.Over)
				break
			} else {
				if testCase.Width+testCase.X >= 100 && testCase.Height+testCase.Y >= 100 {
					//case 23
					draw.Draw(&fragmentBlack, image.Rectangle{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{X: 100 - testCase.X, Y: 100 - testCase.Y},
					}, fragmentRed, image.Point{}, draw.Over)
					break
				} else {
					if testCase.Width+testCase.X >= 100 {
						//case 21
						draw.Draw(&fragmentBlack, image.Rectangle{
							Min: image.Point{X: 0, Y: 0},
							Max: image.Point{X: 100 - testCase.X, Y: testCase.Height},
						}, fragmentRed, image.Point{}, draw.Over)
						break
					} else {
						//case 22
						draw.Draw(&fragmentBlack, image.Rectangle{
							Min: image.Point{X: 0, Y: 0},
							Max: image.Point{X: testCase.Width, Y: 100 - testCase.X},
						}, fragmentRed, image.Point{}, draw.Over)
						break
					}
				}
			}
		}

		url = fmt.Sprintf("/chartas/%s/?x=%d&y=%d&width=%d&height=%d", id, testCase.X, testCase.Y, testCase.Width, testCase.Height)
		req, _ = http.NewRequest("GET", url, nil)
		responseGet := httptest.NewRecorder()
		cs.Router.ServeHTTP(responseGet, req)
		assert.Equal(t, http.StatusOK, responseGet.Code)

		bufLocal := new(bytes.Buffer)
		err = bmp.Encode(bufLocal, &fragmentBlack)
		if err != nil {
			fmt.Println(err)
		}

		eqCode := compareBmp(bufLocal.Bytes(), responseGet.Body.Bytes())
		assert.Equal(t, 0, eqCode)
	}

	for _, testCase := range testCasesBadReq {
		url := fmt.Sprintf("/chartas/%s/?x=%d&y=%d&width=%d&height=%d", id, testCase.X, testCase.Y, testCase.Width, testCase.Height)
		req, _ := http.NewRequest("GET", url, nil)
		response := httptest.NewRecorder()
		cs.Router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}

	_ = fragmentRedFile.Close()

	url = fmt.Sprintf("/chartas/%s/", id)
	req, _ = http.NewRequest("DELETE", url, nil)
	responseDelete := httptest.NewRecorder()
	cs.Router.ServeHTTP(responseDelete, req)
	assert.Equal(t, http.StatusOK, responseDelete.Code)
}

func TestAddFragmentEndpoint(t *testing.T) {
	testCasesOK := []fragmentTest{
		{X: -25, Y: -25, Width: 50, Height: 50},   // 9
		{X: -25, Y: -25, Width: 50, Height: 150},  // 10
		{X: -25, Y: -25, Width: 150, Height: 50},  // 11
		{X: 25, Y: -25, Width: 50, Height: 50},    // 12
		{X: 75, Y: -25, Width: 50, Height: 50},    // 13
		{X: 25, Y: -25, Width: 150, Height: 150},  // 14
		{X: 25, Y: -25, Width: 50, Height: 150},   // 15
		{X: -25, Y: 25, Width: 50, Height: 50},    // 16
		{X: -25, Y: 75, Width: 50, Height: 50},    // 17
		{X: -25, Y: 75, Width: 150, Height: 50},   // 18
		{X: -25, Y: 25, Width: 150, Height: 50},   // 19
		{X: 25, Y: 25, Width: 50, Height: 50},     // 20
		{X: 25, Y: 25, Width: 150, Height: 50},    // 21
		{X: 25, Y: 25, Width: 50, Height: 150},    // 22
		{X: 25, Y: 25, Width: 150, Height: 150},   // 23
		{X: -25, Y: -25, Width: 150, Height: 150}, // 24
	}

	testCasesBadReq := []fragmentTest{
		{X: -100, Y: -100, Width: 50, Height: 50}, // 1
		{X: 40, Y: -60, Width: 50, Height: 50},    // 2
		{X: 150, Y: -100, Width: 50, Height: 50},  // 3
		{X: -100, Y: 50, Width: 50, Height: 50},   // 4
		{X: -100, Y: 150, Width: 50, Height: 50},  // 5
		{X: 150, Y: 40, Width: 50, Height: 50},    // 6
		{X: 150, Y: 150, Width: 50, Height: 50},   // 7
		{X: 40, Y: 150, Width: 50, Height: 50},    // 8
	}

	for _, testCase := range testCasesOK {
		url := fmt.Sprintf("/chartas/?width=%d&height=%d", 100, 100)
		req, _ := http.NewRequest("POST", url, nil)
		responseCreate := httptest.NewRecorder()
		cs.Router.ServeHTTP(responseCreate, req)
		assert.Equal(t, http.StatusCreated, responseCreate.Code)

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(responseCreate.Body)
		if err != nil {
			return
		}
		id := buf.String()

		fragmentImg := createRedImage(testCase.Width, testCase.Height)
		buf = new(bytes.Buffer)
		err = bmp.Encode(buf, fragmentImg)
		if err != nil {
			return
		}

		url = fmt.Sprintf("/chartas/%s/?x=%d&y=%d&width=%d&height=%d", id, testCase.X, testCase.Y, testCase.Width, testCase.Height)
		req, _ = http.NewRequest("POST", url, buf)
		responseAdd := httptest.NewRecorder()
		cs.Router.ServeHTTP(responseAdd, req)
		assert.Equal(t, http.StatusOK, responseAdd.Code)

		redFragment := createRedImage(testCase.Width, testCase.Height)
		chartaImgLocal := createBlackImage(100, 100)

		switch {
		case testCase.X < 0 && testCase.Y < 0:

			//case 24
			if testCase.Width-testCase.X >= 100 && testCase.Height-testCase.Y >= 100 {
				draw.Draw(&chartaImgLocal, image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: 100, Y: 100},
				}, redFragment, image.Point{}, draw.Over)
				break
			}

			if testCase.Height+testCase.Y <= 100 {
				if testCase.Width-testCase.X > 100 {
					//case 11
					draw.Draw(&chartaImgLocal, image.Rectangle{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{X: 100, Y: testCase.Height + testCase.Y},
					}, redFragment, image.Point{}, draw.Over)
					break
				} else {
					//case 9
					draw.Draw(&chartaImgLocal, image.Rectangle{
						Min: image.Point{X: 0, Y: 0},
						Max: image.Point{X: testCase.Width + testCase.X, Y: testCase.Height + testCase.Y},
					}, redFragment, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 10
				draw.Draw(&chartaImgLocal, image.Rectangle{
					Min: image.Point{X: 0, Y: 0},
					Max: image.Point{X: testCase.Width + testCase.X, Y: testCase.Height},
				}, redFragment, image.Point{}, draw.Over)
				break
			}

		case testCase.X >= 0 && testCase.Y < 0:
			if testCase.Height+testCase.Y <= 100 {
				if testCase.X+testCase.Width > 100 {
					//case 13
					draw.Draw(&chartaImgLocal, image.Rectangle{
						Min: image.Point{X: testCase.X, Y: 0},
						Max: image.Point{X: 100, Y: testCase.Height + testCase.Y},
					}, redFragment, image.Point{}, draw.Over)
					break
				} else {
					//case 12
					draw.Draw(&chartaImgLocal, image.Rectangle{
						Min: image.Point{X: testCase.X, Y: 0},
						Max: image.Point{X: testCase.Width + testCase.X, Y: testCase.Height + testCase.Y},
					}, redFragment, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 14 and 15
				draw.Draw(&chartaImgLocal, image.Rectangle{
					Min: image.Point{X: testCase.X, Y: 0},
					Max: image.Point{X: 100, Y: 100},
				}, redFragment, image.Point{}, draw.Over)
				break
			}

		case testCase.X < 0 && testCase.Y >= 0:
			if testCase.Width+testCase.X <= 100 {
				if testCase.Height+testCase.Y >= 100 {
					//case 17
					draw.Draw(&chartaImgLocal, image.Rectangle{
						Min: image.Point{X: 0, Y: testCase.Y},
						Max: image.Point{X: testCase.Width + testCase.X, Y: 100},
					}, redFragment, image.Point{}, draw.Over)
					break
				} else {
					//case 16
					draw.Draw(&chartaImgLocal, image.Rectangle{
						Min: image.Point{X: 0, Y: testCase.Y},
						Max: image.Point{X: testCase.Width + testCase.X, Y: testCase.Height + testCase.Y},
					}, redFragment, image.Point{}, draw.Over)
					break
				}
			} else {
				//case 18 and 19
				draw.Draw(&chartaImgLocal, image.Rectangle{
					Min: image.Point{X: 0, Y: testCase.Y},
					Max: image.Point{X: 100, Y: 100},
				}, redFragment, image.Point{}, draw.Over)
				break
			}

		case testCase.X >= 0 && testCase.Y >= 0:
			//case 20
			if testCase.Width+testCase.X <= 100 && testCase.Height+testCase.Y <= 100 {
				draw.Draw(&chartaImgLocal, image.Rectangle{
					Min: image.Point{X: testCase.X, Y: testCase.Y},
					Max: image.Point{X: testCase.Width + testCase.X, Y: testCase.Height + testCase.Y},
				}, redFragment, image.Point{}, draw.Over)
				break
			} else {
				if testCase.Width+testCase.X >= 100 && testCase.Height+testCase.Y >= 100 {
					//case 23
					draw.Draw(&chartaImgLocal, image.Rectangle{
						Min: image.Point{X: testCase.X, Y: testCase.Y},
						Max: image.Point{X: 100, Y: 100},
					}, redFragment, image.Point{}, draw.Over)
					break
				} else {
					if testCase.Width+testCase.X >= 100 {
						//case 21
						draw.Draw(&chartaImgLocal, image.Rectangle{
							Min: image.Point{X: testCase.X, Y: testCase.Y},
							Max: image.Point{X: 100, Y: testCase.Height + testCase.Y},
						}, redFragment, image.Point{}, draw.Over)
						break
					} else {
						//case 22
						draw.Draw(&chartaImgLocal, image.Rectangle{
							Min: image.Point{X: testCase.X, Y: testCase.Y},
							Max: image.Point{X: testCase.Width + testCase.Y, Y: 100},
						}, redFragment, image.Point{}, draw.Over)
						break
					}
				}
			}
		}

		url = fmt.Sprintf("/chartas/%s/?x=%d&y=%d&width=%d&height=%d", id, 0, 0, 100, 100)
		req, _ = http.NewRequest("GET", url, nil)
		responseGet := httptest.NewRecorder()
		cs.Router.ServeHTTP(responseGet, req)
		assert.Equal(t, http.StatusOK, responseGet.Code)

		bufLocal := new(bytes.Buffer)
		err = bmp.Encode(bufLocal, &chartaImgLocal)
		if err != nil {
			fmt.Println(err)
		}

		eqCode := compareBmp(bufLocal.Bytes(), responseGet.Body.Bytes())
		assert.Equal(t, 0, eqCode)

		url = fmt.Sprintf("/chartas/%s/", id)
		req, _ = http.NewRequest("DELETE", url, nil)
		responseDelete := httptest.NewRecorder()
		cs.Router.ServeHTTP(responseDelete, req)
		assert.Equal(t, http.StatusOK, responseDelete.Code)
	}

	url := fmt.Sprintf("/chartas/?width=%d&height=%d", 100, 100)
	req, _ := http.NewRequest("POST", url, nil)
	responseCreate := httptest.NewRecorder()
	cs.Router.ServeHTTP(responseCreate, req)
	assert.Equal(t, http.StatusCreated, responseCreate.Code)

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(responseCreate.Body)
	if err != nil {
		return
	}
	id := buf.String()

	for _, testCase := range testCasesBadReq {
		url := fmt.Sprintf("/chartas/%s/?x=%d&y=%d&width=%d&height=%d", id, testCase.X, testCase.Y, testCase.Width, testCase.Height)
		req, _ := http.NewRequest("GET", url, nil)
		response := httptest.NewRecorder()
		cs.Router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}

	url = fmt.Sprintf("/chartas/%s/", id)
	req, _ = http.NewRequest("DELETE", url, nil)
	responseDelete := httptest.NewRecorder()
	cs.Router.ServeHTTP(responseDelete, req)
	assert.Equal(t, http.StatusOK, responseDelete.Code)

	filename := fmt.Sprintf("chartas/%s.png", id)
	err = os.Remove(filename)
	if err != nil {
		return
	}

}

func TestDeleteFragmentEndpoint(t *testing.T) {
	url := fmt.Sprintf("/chartas/?width=10&height=10")

	req, _ := http.NewRequest("POST", url, nil)
	response := httptest.NewRecorder()
	cs.Router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusCreated, response.Code)

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(response.Body)
	if err != nil {
		return
	}
	id := buf.String()

	url = fmt.Sprintf("/chartas/%s/", id)
	responseDelete := httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", url, nil)
	cs.Router.ServeHTTP(responseDelete, req)
	assert.Equal(t, http.StatusOK, responseDelete.Code)
	filename := fmt.Sprintf("chartas/%s.png", id)
	_, err = os.Open(filename)

	errString := fmt.Sprintf("open chartas/%s.png: no such file or directory", id)
	assert.Equal(t, errString, err.Error())

}

func createRedImage(width, height int) *image.NRGBA {
	buf := make([]uint8, height*width*4)
	var i int64
	for i = 3; i <= int64(height*width*4); i += 4 {
		buf[i] = 255
		buf[i-3] = 255
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	img.Pix = buf
	img.Stride = 4 * width

	return img
}

func compareBmp(bytes1, bytes2 []byte) int {
	if len(bytes1) != len(bytes2) {
		return -1
	}

	for i := 0; i < len(bytes1); i++ {
		if bytes1[i] != bytes2[i] {
			return i
		}
	}

	return 0
}
