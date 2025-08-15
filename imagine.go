package imagine

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
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
		p.Hasher = SHA256Hasher()
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
	fmt.Printf("[Imagine] Get called for filename: %s\n", filename)
	
	cacheKey, err := i.cacheKey(filename, params)
	if err != nil {
		return nil, errors.Trace(err)
	}
	fmt.Printf("[Imagine] Cache key: %s\n", cacheKey)

	// try to grab the image from cache
	var image []byte
	image, found, err := i.params.Cache.Get(cacheKey)
	if err != nil {
		fmt.Printf("[Imagine] Cache.Get error: %v (found: %v)\n", err, found)
		// If it's just a cache miss (not found), that's OK - continue to storage
		if errors.Is(err, ErrKeyNotFound) {
			fmt.Printf("[Imagine] Cache miss is expected, checking storage\n")
			err = nil // Clear the error for cache miss
		} else {
			return nil, errors.Trace(err)
		}
	} else if found {
		fmt.Printf("[Imagine] Found in cache, returning cached image\n")
		bi := bimg.NewImage(image)
		return &ProcessedImage{Image: bi.Image(), Type: bi.Type()}, nil
	}

	// get it from storage if not in cache
	fmt.Printf("[Imagine] Not in cache, checking storage for filename: %s\n", filename)
	image, found, err = i.params.Storage.Get(filename)
	if err != nil {
		fmt.Printf("[Imagine] Storage.Get error: %v (found: %v), Error type: %T\n", err, found, err)
		// If it's just not found, return a placeholder
		// Check both the error value and the error message
		if err == ErrKeyNotFound || errors.Is(err, ErrKeyNotFound) || err.Error() == "key not found" {
			fmt.Printf("[Imagine] Image not found, generating placeholder\n")
			return i.getPlaceholderImage(params)
		}
		return nil, errors.Trace(err)
	} else if !found {
		fmt.Printf("[Imagine] Image not found in storage, returning placeholder\n")
		return i.getPlaceholderImage(params)
	}
	fmt.Printf("[Imagine] Found in storage, size: %d bytes\n", len(image))

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
	fmt.Printf("[Imagine] Upload called with data size: %d bytes (%.2f MB)\n", len(data), float64(len(data))/1024/1024)
	
	if isValid := validateImage(data); !isValid {
		contentType := http.DetectContentType(data)
		fmt.Printf("[Imagine] Invalid image type. Detected content type: %s\n", contentType)
		return "", errors.New("invalid image type")
	}
	fmt.Printf("[Imagine] Image validation passed\n")

	// Auto-orient and optimize the image before storing
	img := bimg.NewImage(data)
	
	// Get metadata to check orientation and size
	metadata, err := img.Metadata()
	if err == nil {
		fmt.Printf("[Imagine] Original image: %dx%d, orientation: %d\n", 
			metadata.Size.Width, metadata.Size.Height, metadata.Orientation)
		
		// Build options for optimization
		options := bimg.Options{
			StripMetadata: true, // Remove metadata
		}
		
		// Auto-rotate if needed
		if metadata.Orientation > 1 {
			fmt.Printf("[Imagine] Auto-rotating image\n")
			data, err = img.AutoRotate()
			if err != nil {
				fmt.Printf("[Imagine] Warning: Failed to auto-rotate: %v\n", err)
			} else {
				img = bimg.NewImage(data) // Recreate with rotated data
			}
		}
		
		// If image is very large (>4096px), resize it for storage optimization
		if metadata.Size.Width > 4096 || metadata.Size.Height > 4096 {
			fmt.Printf("[Imagine] Image is very large, resizing to max 4096px for storage\n")
			if metadata.Size.Width > metadata.Size.Height {
				options.Width = 4096
			} else {
				options.Height = 4096
			}
		}
		
		// Ensure file is under 10MB
		targetSize := 10 * 1024 * 1024 // 10MB
		currentSize := len(data)
		
		if currentSize > targetSize {
			fmt.Printf("[Imagine] File size %.2f MB exceeds 10MB limit, optimizing...\n", float64(currentSize)/1024/1024)
			
			// Calculate dimension reduction needed
			reductionFactor := math.Sqrt(float64(targetSize) / float64(currentSize))
			
			// Apply dimension reduction if not already set
			if options.Width == 0 && options.Height == 0 {
				newWidth := int(float64(metadata.Size.Width) * reductionFactor)
				newHeight := int(float64(metadata.Size.Height) * reductionFactor)
				
				// Cap at 4096px max
				if newWidth > 4096 {
					newWidth = 4096
				}
				if newHeight > 4096 {
					newHeight = 4096
				}
				
				if metadata.Size.Width > metadata.Size.Height {
					options.Width = newWidth
				} else {
					options.Height = newHeight
				}
				
				fmt.Printf("[Imagine] Resizing to approximately %dx%d\n", newWidth, newHeight)
			}
			
			// Use JPEG compression for large files (unless PNG with transparency)
			if metadata.Type != "png" {
				options.Type = bimg.JPEG
				options.Quality = 92 // Good quality but more compression
			}
		} else if len(data) > 5*1024*1024 && metadata.Type != "png" { 
			// For files between 5-10MB, still optimize
			fmt.Printf("[Imagine] File over 5MB, applying compression\n")
			options.Type = bimg.JPEG
			options.Quality = 95
		}
		
		// Apply optimizations if any were set
		if options.Width > 0 || options.Height > 0 || options.Type > 0 {
			data, err = img.Process(options)
			if err != nil {
				fmt.Printf("[Imagine] Warning: Failed to optimize: %v\n", err)
				// Continue with original data if optimization fails
			} else {
				fmt.Printf("[Imagine] Optimized to %d bytes (%.2f MB)\n", 
					len(data), float64(len(data))/1024/1024)
			}
		}
	}
	
	// use the file hash as the filename (after processing)
	filename, err := i.params.Hasher.Hash(data)
	if err != nil {
		fmt.Printf("[Imagine] Failed to hash data: %v\n", err)
		return "", errors.Trace(err)
	}
	fmt.Printf("[Imagine] Generated hash filename: %s\n", filename)

	err = i.params.Storage.Set(filename, data)
	if err != nil {
		fmt.Printf("[Imagine] Failed to store image: %v\n", err)
		return "", errors.Trace(err)
	}
	fmt.Printf("[Imagine] Successfully stored image with filename: %s\n", filename)

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
	
	// Preset for common configurations: thumb, small, medium, large, hero
	Preset string
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
	
	// Handle presets
	if queryValues.Has("preset") {
		p.Preset = queryValues.Get("preset")
		// Apply preset defaults (will be overridden by specific params if provided)
		switch p.Preset {
		case "thumb":
			// Small thumbnail - 150x150 square, lower quality
			if p.Width == 0 && p.Height == 0 {
				p.Width = 150
				p.Height = 150
				p.Fit = "cover"
			}
			if p.Quality == 0 {
				p.Quality = 80
			}
			if p.Format == "" {
				p.Format = "webp"
			}
		case "small":
			// Small image - 400px wide
			if p.Width == 0 {
				p.Width = 400
			}
			if p.Quality == 0 {
				p.Quality = 85
			}
			if p.Format == "" {
				p.Format = "webp"
			}
		case "medium":
			// Medium image - 800px wide
			if p.Width == 0 {
				p.Width = 800
			}
			if p.Quality == 0 {
				p.Quality = 85
			}
			if p.Format == "" {
				p.Format = "webp"
			}
		case "large":
			// Large image - 1200px wide
			if p.Width == 0 {
				p.Width = 1200
			}
			if p.Quality == 0 {
				p.Quality = 85
			}
			if p.Format == "" {
				p.Format = "webp"
			}
		case "hero":
			// Hero/banner image - 1920px wide, higher quality
			if p.Width == 0 {
				p.Width = 1920
			}
			if p.Quality == 0 {
				p.Quality = 90
			}
			if p.Format == "" {
				p.Format = "webp"
			}
		case "placeholder":
			// Tiny blurred placeholder - 20px wide, very low quality
			if p.Width == 0 {
				p.Width = 20
			}
			if p.Quality == 0 {
				p.Quality = 20
			}
			if p.Blur == 0 {
				p.Blur = 5
			}
			if p.Format == "" {
				p.Format = "webp"
			}
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
	// First auto-rotate the image if needed
	img := bimg.NewImage(image)
	metadata, err := img.Metadata()
	if err == nil && metadata.Orientation > 1 {
		// Auto-rotate based on EXIF orientation
		image, err = img.AutoRotate()
		if err != nil {
			// Log but continue with original image
			fmt.Printf("[Imagine] Warning: Failed to auto-rotate in processImage: %v\n", err)
		}
		// Recreate img with rotated data
		img = bimg.NewImage(image)
	}
	
	options := bimg.Options{}
	
	// Apply smart web defaults if no specific params are provided
	hasTransformations := params.Width > 0 || params.Height > 0 || params.Thumbnail > 0 || 
		params.Format != "" || params.Quality > 0 || params.Fit != ""
	
	if !hasTransformations {
		// Get image dimensions to apply smart defaults
		size, _ := img.Size()
		
		// Apply web-optimized defaults
		// Limit max dimension to 2048px for web display
		if size.Width > 2048 || size.Height > 2048 {
			if size.Width > size.Height {
				options.Width = 2048
			} else {
				options.Height = 2048
			}
		}
		
		// Default to WebP format for better compression
		options.Type = bimg.WEBP
		
		// Default quality of 85 for good balance
		options.Quality = 85
		
		// Enable strip to remove metadata for smaller files
		options.StripMetadata = true
		
		fmt.Printf("[Imagine] Applying web defaults: max 2048px, WebP, quality 85\n")
	}

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

	// Quality (default to 85 if not specified)
	if params.Quality > 0 {
		options.Quality = params.Quality
	} else if !hasTransformations {
		// Already set above
	} else {
		// Default quality for any transformation
		options.Quality = 85
	}
	
	// Always strip metadata to reduce file size
	options.StripMetadata = true

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

// getPlaceholderImage returns a placeholder image when the requested image is not found
func (i *Imagine) getPlaceholderImage(params *ImageParams) (*ProcessedImage, error) {
	// Determine size for placeholder
	width := 400
	height := 300
	if params != nil {
		if params.Width > 0 {
			width = params.Width
		}
		if params.Height > 0 {
			height = params.Height
		}
		// Ensure minimum size for visibility
		if width < 100 {
			width = 100
		}
		if height < 100 {
			height = 100
		}
	}
	
	// Create a minimal valid gray PNG (10x10 pixels)
	// This is more likely to work with bimg than a 1x1
	// Using a pre-generated 10x10 gray PNG
	baseImage := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x0a,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x02, 0x50, 0x58, 0xea, 0x00, 0x00, 0x00,
		0x15, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0xfa, 0xcf, 0xc0, 0xc0,
		0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x00, 0x00, 0x00, 0xff, 0x00, 0x01,
		0x21, 0xd5, 0x05, 0xfe, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
		0xae, 0x42, 0x60, 0x82,
	}
	
	// Create bimg image from base
	img := bimg.NewImage(baseImage)
	
	// Resize to desired dimensions with gray background
	options := bimg.Options{
		Width:      width,
		Height:     height,
		Type:       bimg.PNG,
		Background: bimg.Color{R: 240, G: 240, B: 240}, // Light gray
		Force:      true, // Force exact dimensions
	}
	
	processed, err := img.Process(options)
	if err != nil {
		fmt.Printf("[Imagine] Failed to create placeholder image: %v\n", err)
		// Return an error but don't crash - EditorJS will handle it
		return nil, errors.Trace(err)
	}
	
	fmt.Printf("[Imagine] Created placeholder image: %dx%d\n", width, height)
	
	return &ProcessedImage{
		Image: processed,
		Type:  "image/png",
	}, nil
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
