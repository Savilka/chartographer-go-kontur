package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
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
	cs.Initialize("", "test.db")
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

		filename := fmt.Sprintf("chartas/%s.png", id)
		file, err := os.Open(filename)
		if err != nil {
			return
		}

		img, err := png.Decode(file)
		if err != nil {
			return
		}

		_ = file.Close()

		assert.Equal(t, testCase.Width, img.Bounds().Dx())
		assert.Equal(t, testCase.Height, img.Bounds().Dy())

		err = os.Remove(filename)
		if err != nil {
			return
		}

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

//func TestGetFragmentEndpoint(t *testing.T) {
//	testCasesOK := []fragmentTest{
//		{X: -25, Y: -25, Width: 50, Height: 50},   // 9
//		{X: -25, Y: -25, Width: 50, Height: 150},  // 10
//		{X: -25, Y: -25, Width: 150, Height: 50},  // 11
//		{X: 25, Y: -25, Width: 50, Height: 50},    // 12
//		{X: 75, Y: -25, Width: 50, Height: 50},    // 13
//		{X: 25, Y: -25, Width: 150, Height: 150},  // 14
//		{X: 25, Y: -25, Width: 50, Height: 150},   // 15
//		{X: -25, Y: 25, Width: 50, Height: 50},    // 16
//		{X: -25, Y: 75, Width: 50, Height: 50},    // 17
//		{X: -25, Y: 75, Width: 150, Height: 50},   // 18
//		{X: -25, Y: 25, Width: 150, Height: 50},   // 19
//		{X: 25, Y: 25, Width: 50, Height: 50},     // 20
//		{X: 25, Y: 25, Width: 150, Height: 50},    // 21
//		{X: 25, Y: 25, Width: 50, Height: 150},    // 22
//		{X: 25, Y: 25, Width: 150, Height: 150},   // 23
//		{X: -25, Y: -25, Width: 150, Height: 150}, // 24
//	}
//
//	testCasesBadReq := []fragmentTest{
//		{X: -100, Y: -100, Width: 50, Height: 50}, // 1
//		{X: 40, Y: -60, Width: 50, Height: 50},    // 2
//		{X: 150, Y: -100, Width: 50, Height: 50},  // 3
//		{X: -100, Y: 50, Width: 50, Height: 50},   // 4
//		{X: -100, Y: 150, Width: 50, Height: 50},  // 5
//		{X: 150, Y: 40, Width: 50, Height: 50},    // 6
//		{X: 150, Y: 150, Width: 50, Height: 50},   // 7
//		{X: 40, Y: 150, Width: 50, Height: 50},    // 8
//	}
//
//	for _, testCase := range testCasesOK {
//
//		fragmentBg := createBlackImage(testCase.Width, testCase.Height)
//	}
//
//	for _, testCase := range testCasesBadReq {
//
//		fragmentBg := createBlackImage(testCase.Width, testCase.Height)
//	}
//}

func TestAddFragmentEndpoint(t *testing.T) {

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

	errString := fmt.Sprintf("open chartas/%s.png: The system cannot find the file specified.", id)
	assert.Equal(t, errString, err.Error())

}
