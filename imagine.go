package imagine

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/h2non/bimg"
	"github.com/juju/errors"
)

var ErrImageNotFound = errors.New("image not found")

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

	params, err := i.ParamsFromQueryString(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pi, err := i.Get(slug, params)
	if err != nil && errors.Cause(err) == ErrImageNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", pi.Type)
	w.Write(pi.Image)
}

// ProcessedImage is the result of processing an image
type ProcessedImage struct {
	Type  string
	Image []byte
}

// Get is the main entry point for the Imagine application. It returns the
// image as an array of bytes
func (i *Imagine) Get(filename string, params *ImageParams) (*ProcessedImage, error) {
	cacheKey, err := i.cacheKey(filename, params)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// try to grab the image from cache
	var image []byte
	image, found, err := i.params.Cache.Get(cacheKey)
	if err != nil {
		return nil, errors.Trace(err)
	} else if found {
		bi := bimg.NewImage(image)
		return &ProcessedImage{Image: bi.Image(), Type: bi.Type()}, nil
	}

	// get it from storage if not in cache
	image, found, err = i.params.Storage.Get(filename)
	if err != nil {
		return nil, errors.Trace(err)
	} else if !found {
		return nil, ErrImageNotFound
	}

	// apply the requested transformations
	processedImage, err := i.processImage(image, params)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// store the processed image in cache
	err = i.params.Cache.Set(cacheKey, processedImage.Image())
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &ProcessedImage{
		Type:  processedImage.Type(),
		Image: processedImage.Image(),
	}, nil
}

func (i *Imagine) cacheKey(filename string, params *ImageParams) (string, error) {
	paramsCacheKey, err := params.CacheKey(i.params.Hasher)
	if err != nil {
		return "", errors.Trace(err)
	}

	return fmt.Sprintf("%s%s", filename, paramsCacheKey), nil
}

func (i *Imagine) Upload(data []byte) (string, error) {
	// use the file hash as the filename
	filename, err := i.params.Hasher.Hash(data)
	if err != nil {
		return "", errors.Trace(err)
	}

	if isValid := validateImage(data); !isValid {
		return "", errors.New("invalid image type")
	}

	err = i.params.Storage.Set(filename, data)
	if err != nil {
		return "", errors.Trace(err)
	}

	return filename, nil
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

	filename, err := i.Upload(imgBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(filename))
}

// Image params are the requested params to modify an image when retriving it
type ImageParams struct {
	Width, Height int

	// Quality is the quality of the image to be returned
	Quality int

	// Format is the format of the image to be returned
	Format string

	// Thumbnail is the size of the thumbnail to be returned
	Thumbnail int
}

// create a cache key from the image params
func (ip *ImageParams) CacheKey(h Hasher) (string, error) {
	hash, err := h.Hash([]byte(fmt.Sprintf("%v", ip)))
	if err != nil {
		return "", errors.Trace(err)
	}

	return hash, nil
}

// ParamsFromQueryString returns an ImageParams given a query string
func (i *Imagine) ParamsFromQueryString(query string) (*ImageParams, error) {
	p := ImageParams{}
	u, err := url.Parse(query)
	if err != nil {
		return nil, errors.Trace(err)
	}
	queryValues := u.Query()

	// TODO: add support for custom query params
	if queryValues.Has("w") {
		width, err := strconv.Atoi(queryValues.Get("w"))
		if err != nil {
			return nil, errors.Trace(err)
		}
		p.Width = width
	}

	if queryValues.Has("h") {
		height, err := strconv.Atoi(queryValues.Get("h"))
		if err != nil {
			return nil, errors.Trace(err)
		}
		p.Height = height
	}

	if queryValues.Has("format") {
		height, err := strconv.Atoi(queryValues.Get("height"))
		if err != nil {
			return nil, errors.Trace(err)
		}
		p.Height = height
	}

	return &p, nil
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

// processImage applies the requested transformations to the image via the supplied params
func (i *Imagine) processImage(image []byte, params *ImageParams) (*bimg.Image, error) {
	var err error
	if params.Width > 0 && params.Height > 0 {
		image, err = bimg.NewImage(image).Resize(params.Width, params.Height)
		if err != nil {
			return nil, errors.Trace(err)
		}
	} else if params.Height > 0 {
		image, err = bimg.NewImage(image).CropByHeight(params.Height)
		if err != nil {
			return nil, errors.Trace(err)
		}
	} else if params.Width > 0 {
		image, err = bimg.NewImage(image).CropByHeight(params.Width)
		if err != nil {
			return nil, errors.Trace(err)
		}
	} else if params.Thumbnail > 0 {
		image, err = bimg.NewImage(image).Thumbnail(params.Thumbnail)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	if params.Format != "" {
		var format bimg.ImageType
		switch params.Format {
		case "png":
			format = bimg.PDF
		case "webp":
			format = bimg.WEBP
		}

		image, err = bimg.NewImage(image).Convert(format)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return bimg.NewImage(image), nil
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
