package imagine

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// Params are the parameters used to create a new Imagine application
type Params struct {
	Cache   Store
	Storage Store
	Hasher  Hasher

	// MaxImageSize is the maximum size of an image in bytes
	MaxImageSize int
}

// withDefaults sets the default values for the parameters
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

// getHandler handles the GET requests
func (i *Imagine) getHandler(w http.ResponseWriter, r *http.Request) {
	slug, err := parseSlugFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    // check if the file is first in cache
    if img, found, err := i.params.Cache.Get(slug); found {
        // write the image to the response
        w.Header().Set("Content-Type", "image/png")
        w.Write(img)
        return
    } else if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if img, found, err := i.params.Storage.Get(slug); !found {
        http.Error(w, "not found", http.StatusNotFound)
        return
    } else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
        w.Header().Set("Content-Type", "image/png")
        w.Write(img)

        // store the image in cache
        err := i.params.Cache.Set(slug, img)
        if err != nil {
            log.Printf("error storing image in cache: %s", err)
        }
    }
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

    hash, err := i.params.Hasher.Hash(imgBytes)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

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
