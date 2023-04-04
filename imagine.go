package imagine

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/risico/imagine/src/cache"
	"github.com/risico/imagine/src/storage"
)

type Params struct {
	Cache   cache.Cacher
	Storage storage.Storage
}

// Imagine is our main application struct
type Imagine struct {
	params Params
}

// UploadHandler handles the upload of images
func (i *Imagine) UploadHandlerFunc() http.HandlerFunc {
	return http.HandlerFunc(i.uploadHandlerFunc)
}

// ProcessHandler handles the generation of images
func (i *Imagine) GetHandlerFunc() http.HandlerFunc {
	return http.HandlerFunc(i.getHandler)
}

// New creates a new Imagine application
func New(params Params) (*Imagine, error) {
	return &Imagine{
		params: params,
	}, nil
}

func (i *Imagine) getHandler(w http.ResponseWriter, r *http.Request) {
	slug, err := parseSlugFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	img, err := i.params.Storage.Get(slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("slug", slug)
	fmt.Println("img", string(img))
}

func (i *Imagine) uploadHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	imgReader := io.LimitReader(file, 1000000)
	imgBytes, err := ioutil.ReadAll(imgReader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hash := getMD5Hash(imgBytes)
	fmt.Println("hash", hash)

	err = i.params.Storage.Set(hash, imgBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

var matcher = regexp.MustCompile(`[a-zA-Z0-9]+\.[a-zA-Z]{3,4}$`)

func parseSlugFromPath(path string) (string, error) {
	s := matcher.FindString(path)
	if s == "" {
		return "", fmt.Errorf("no slug found in path: %s", path)
	}

	return s, nil
}

// generate the md5 hash of the image
func getMD5Hash(img []byte) string {
	hash := md5.Sum(img)
	return hex.EncodeToString(hash[:])
}
