package imagine

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/risico/imagine/src/cache"
	"github.com/risico/imagine/src/storage"
)

type Params struct {
	Cache   cache.Cacher
	Storage storage.FS
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

	fmt.Println("slug", slug)
}

func (i *Imagine) uploadHandlerFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("r.URL.Path", r.URL.Path)
}

var matcher = regexp.MustCompile(`[a-zA-Z0-9]+\.[a-zA-Z]{3,4}$`)

func parseSlugFromPath(path string) (string, error) {
	s := matcher.FindString(path)
	if s == "" {
		return "", fmt.Errorf("no slug found in path: %s", path)
	}

	return s, nil
}
