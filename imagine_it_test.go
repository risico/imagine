package imagine_test

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/risico/imagine"
	"github.com/risico/imagine/src/cache"
	"github.com/risico/imagine/src/storage"
)

func TestUploadHandler(t *testing.T) {
	i, err := imagine.New(imagine.Params{
		Storage: storage.NewLocalStorage("/tmp"),
		Cache:   &cache.InMemoryCache{},
	})
	assert.NoError(t, err)

	// Set up a pipe to avoid buffering
	pr, pw := io.Pipe()
	// This writer is going to transform
	// what we pass to it to multipart form data
	// and write it to our io.Pipe
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()
		// We create the form data field 'fileupload'
		// which returns another writer to write the actual file
		part, err := writer.CreateFormFile("file", "someimg.png")
		if err != nil {
			t.Error(err)
		}

		img := createImage()

		// Encode() takes an io.Writer.
		// We pass the multipart field
		// 'fileupload' that we defined
		// earlier which, in turn, writes
		// to our io.Pipe
		err = png.Encode(part, img)
		if err != nil {
			t.Error(err)
		}
	}()

	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	handler := i.UploadHandlerFunc()
	handler.ServeHTTP(response, request)

	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, string(body))
}

func TestGetImagineHandler(t *testing.T) {
	t.Skip("not implemented yet")
	i, err := imagine.New(imagine.Params{
		Storage: storage.NewLocalStorage("/tmp"),
		Cache:   &cache.InMemoryCache{},
	})
	assert.NoError(t, err)

	ts := httptest.NewUnstartedServer(i.GetHandlerFunc())
	ts.Start()

	resp, err := ts.Client().Get(fmt.Sprintf("%s/imagine/someiage.png", ts.URL))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func createImage() image.Image {
	width := 200
	height := 100

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}
	return img
}
