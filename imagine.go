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
)

type Params struct {
	Cache   cache.Cacher
	Storage Storage
	Hasher  Hasher

	// MaxImageSize is the maximum size of an image in bytes
	MaxImageSize int
}

func (p *Params) withDefaults() {
	if p.MaxImageSize == 0 {
		p.MaxImageSize = 1024 * 1024
	}

	if p.Hasher == nil {
		p.Hasher = MD5Hasher()
	}
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
	params.withDefaults()
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
	// nothing to do unless we deal with a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: maybe make this an array so we can support bulk uploads
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	imgBytes, err := ioutil.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isValid := validateImage(imgBytes); !isValid {
		http.Error(w, "invalid image type", http.StatusBadRequest)
		return
	}

	hash := getMD5Hash(imgBytes)

	err = i.params.Storage.Set(hash, imgBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	// TODO: figure out if I should return JSON or simple text
	w.Write([]byte(hash))
}

// pathMatcher matches any path that has some chars and ends in an extension.
var pathMatcher = regexp.MustCompile(`[a-zA-Z0-9]{32}\.[a-zA-Z]{3,4}$`)

// parseSlugFromPath parses the slug from the path
func parseSlugFromPath(path string) (string, error) {
	s := pathMatcher.FindString(path)
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

// validateImage checks if the image is a valid image based on
// the content type.
func validateImage(img []byte) bool {
	ct := http.DetectContentType(img)
	switch ct {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	}

	return false
}
