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

	// Quality is the quality of the image to be returned (1-100)
	Quality int

	// Format is the format of the image to be returned
	Format string

	// Thumbnail is the size of the thumbnail to be returned
	Thumbnail int

	// Fit mode: cover, contain, fill, inside, outside
	Fit string

	// Rotation angle in degrees (0, 90, 180, 270)
	Rotate int

	// Flip: h (horizontal), v (vertical), both
	Flip string

	// Blur sigma value (0.3 to 1000)
	Blur float64

	// Sharpen sigma value  
	Sharpen float64

	// Convert to grayscale
	Grayscale bool

	// Gravity for smart cropping: center, north, south, east, west, etc.
	Gravity string
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
		p.Format = queryValues.Get("format")
	}

	if queryValues.Has("q") || queryValues.Has("quality") {
		qKey := "q"
		if !queryValues.Has("q") {
			qKey = "quality"
		}
		quality, err := strconv.Atoi(queryValues.Get(qKey))
		if err != nil {
			return nil, errors.Trace(err)
		}
		if quality < 1 || quality > 100 {
			return nil, errors.New("quality must be between 1 and 100")
		}
		p.Quality = quality
	}

	if queryValues.Has("thumbnail") {
		thumbnail, err := strconv.Atoi(queryValues.Get("thumbnail"))
		if err != nil {
			return nil, errors.Trace(err)
		}
		p.Thumbnail = thumbnail
	}

	if queryValues.Has("fit") {
		fit := queryValues.Get("fit")
		switch fit {
		case "cover", "contain", "fill", "inside", "outside":
			p.Fit = fit
		default:
			return nil, errors.New("invalid fit mode: must be cover, contain, fill, inside, or outside")
		}
	}

	if queryValues.Has("rotate") {
		rotate, err := strconv.Atoi(queryValues.Get("rotate"))
		if err != nil {
			return nil, errors.Trace(err)
		}
		if rotate != 0 && rotate != 90 && rotate != 180 && rotate != 270 {
			return nil, errors.New("rotate must be 0, 90, 180, or 270")
		}
		p.Rotate = rotate
	}

	if queryValues.Has("flip") {
		flip := queryValues.Get("flip")
		switch flip {
		case "h", "v", "both":
			p.Flip = flip
		default:
			return nil, errors.New("flip must be h, v, or both")
		}
	}

	if queryValues.Has("blur") {
		blur, err := strconv.ParseFloat(queryValues.Get("blur"), 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		if blur < 0.3 || blur > 1000 {
			return nil, errors.New("blur must be between 0.3 and 1000")
		}
		p.Blur = blur
	}

	if queryValues.Has("sharpen") {
		sharpen, err := strconv.ParseFloat(queryValues.Get("sharpen"), 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		p.Sharpen = sharpen
	}

	if queryValues.Has("grayscale") || queryValues.Has("greyscale") {
		p.Grayscale = true
	}

	if queryValues.Has("gravity") {
		gravity := queryValues.Get("gravity")
		switch gravity {
		case "center", "centre", "north", "south", "east", "west", "smart":
			p.Gravity = gravity
		case "northeast", "northwest", "southeast", "southwest":
			// These will be mapped to smart crop since bimg doesn't support them directly
			p.Gravity = gravity
		default:
			return nil, errors.New("invalid gravity value")
		}
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
	options := bimg.Options{}
	var err error

	// Handle sizing based on fit mode
	if params.Fit != "" && params.Width > 0 && params.Height > 0 {
		switch params.Fit {
		case "cover":
			options.Width = params.Width
			options.Height = params.Height
			options.Crop = true
			if params.Gravity != "" {
				options.Gravity = getGravity(params.Gravity)
			}
		case "contain", "inside":
			options.Width = params.Width
			options.Height = params.Height
			options.Embed = true
		case "fill":
			options.Width = params.Width
			options.Height = params.Height
			options.Force = true
		case "outside":
			options.Width = params.Width
			options.Height = params.Height
			options.Enlarge = true
		}
	} else if params.Thumbnail > 0 {
		options.Width = params.Thumbnail
		options.Height = params.Thumbnail
		options.Crop = true
		if params.Gravity != "" {
			options.Gravity = getGravity(params.Gravity)
		}
	} else {
		// Original resize logic
		if params.Width > 0 && params.Height > 0 {
			options.Width = params.Width
			options.Height = params.Height
		} else if params.Height > 0 {
			options.Height = params.Height
		} else if params.Width > 0 {
			options.Width = params.Width
		}
	}

	// Rotation
	if params.Rotate > 0 {
		switch params.Rotate {
		case 90:
			options.Rotate = bimg.D90
		case 180:
			options.Rotate = bimg.D180
		case 270:
			options.Rotate = bimg.D270
		}
	}

	// Flip
	if params.Flip != "" {
		switch params.Flip {
		case "h":
			options.Flip = true
		case "v":
			options.Flop = true
		case "both":
			options.Flip = true
			options.Flop = true
		}
	}

	// Effects
	if params.Blur > 0 {
		options.GaussianBlur = bimg.GaussianBlur{
			Sigma: params.Blur,
		}
	}

	if params.Sharpen > 0 {
		options.Sharpen = bimg.Sharpen{
			Radius: int(params.Sharpen),
		}
	}

	if params.Grayscale {
		options.Interpretation = bimg.InterpretationBW
	}

	// Quality
	if params.Quality > 0 {
		options.Quality = params.Quality
	}

	// Format conversion
	if params.Format != "" {
		switch params.Format {
		case "jpeg", "jpg":
			options.Type = bimg.JPEG
		case "png":
			options.Type = bimg.PNG
		case "webp":
			options.Type = bimg.WEBP
		case "gif":
			options.Type = bimg.GIF
		case "tiff":
			options.Type = bimg.TIFF
		case "avif":
			options.Type = bimg.AVIF
		default:
			return nil, errors.New("unsupported format: " + params.Format)
		}
	}

	// Process the image with all options
	image, err = bimg.NewImage(image).Process(options)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return bimg.NewImage(image), nil
}

// getGravity converts string gravity to bimg.Gravity
func getGravity(gravity string) bimg.Gravity {
	switch gravity {
	case "north":
		return bimg.GravityNorth
	case "south":
		return bimg.GravitySouth
	case "east":
		return bimg.GravityEast
	case "west":
		return bimg.GravityWest
	case "smart":
		return bimg.GravitySmart
	case "center", "centre":
		return bimg.GravityCentre
	default:
		// For compound directions, default to smart crop or center
		// since bimg doesn't support all compass points
		if gravity == "northeast" || gravity == "northwest" || 
		   gravity == "southeast" || gravity == "southwest" {
			return bimg.GravitySmart
		}
		return bimg.GravityCentre
	}
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
