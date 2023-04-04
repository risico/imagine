package imagine_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

	ts := httptest.NewUnstartedServer(i.UploadHandlerFunc())
	ts.Start()

	testImageBytes := []byte("test image bytes")
	testImageReader := bytes.NewReader(testImageBytes)

	resp, err := ts.Client().Post(ts.URL, "image/jpeg", testImageReader)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "someimage.png", string(body))
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
