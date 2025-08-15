# Imagine

[![Go Reference](https://pkg.go.dev/badge/github.com/risico/imagine.svg)](https://pkg.go.dev/github.com/risico/imagine)
[![Go Report Card](https://goreportcard.com/badge/github.com/risico/imagine)](https://goreportcard.com/report/github.com/risico/imagine)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, plug-and-play image processing service for Go applications. Transform images on-the-fly using simple URL parameters.

![Imagine Banner](https://i.imgur.com/xJPvrPI.png)

## ✨ Features

- **🚀 High Performance** - Built on [libvips](https://github.com/libvips/libvips) through [bimg](https://github.com/h2non/bimg) for blazing-fast image operations
- **🔌 Plug & Play** - Works as a standalone server or embedded library
- **🎨 Rich Transformations** - Resize, crop, rotate, blur, sharpen, format conversion, and more
- **💾 Flexible Storage** - Multiple storage backends (Memory, Local, Redis, SQLite, BoltDB)
- **⚡ Smart Caching** - Multi-tier caching with configurable TTL
- **🔗 URL-Based API** - Transform images using simple query parameters
- **📦 Zero Configuration** - Sensible defaults that just work

## 📚 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [As a Library](#as-a-library)
  - [As a Standalone Server](#as-a-standalone-server)
- [URL Parameters](#url-parameters)
- [Storage Backends](#storage-backends)
- [Examples](#examples)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [License](#license)

## 🔧 Installation

### Prerequisites

Imagine requires libvips to be installed on your system:

**macOS:**
```bash
brew install vips
```

**Ubuntu/Debian:**
```bash
sudo apt-get install libvips-dev
```

**RHEL/CentOS:**
```bash
sudo yum install vips-devel
```

### Install Package

```bash
go get github.com/risico/imagine
```

## 🚀 Quick Start

### As a Library

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/risico/imagine"
)

func main() {
    // Create imagine instance with defaults
    img, err := imagine.New(imagine.Params{
        Storage: imagine.NewLocalStorage(imagine.LocalStorageParams{
            Path: "./images",
        }),
        Cache: imagine.NewMemoryStorage(imagine.MemoryStoreParams{}),
    })
    if err != nil {
        log.Fatal(err)
    }

    // Register HTTP handlers
    http.HandleFunc("/upload", img.UploadHandlerFunc())
    http.HandleFunc("/images/", img.GetHandlerFunc())
    
    log.Println("Server running on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### As a Standalone Server

```bash
# Install the CLI
go install github.com/risico/imagine/cmd@latest

# Start the server
imagine start --hostname localhost --port 8080
```

## 📖 Usage

### As a Library

#### Basic Integration

```go
import "github.com/risico/imagine"

// Initialize with custom configuration
img, err := imagine.New(imagine.Params{
    Storage: imagine.NewLocalStorage(imagine.LocalStorageParams{
        Path: "/var/images",
    }),
    Cache: imagine.NewRedisStorage(imagine.RedisStoreParams{
        Addr: "localhost:6379",
        TTL:  24 * time.Hour,
    }),
    MaxImageSize: 10 * 1024 * 1024, // 10MB
})
```

#### With Gin Framework

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/risico/imagine"
)

func main() {
    r := gin.Default()
    
    img, _ := imagine.New(imagine.Params{
        Storage: imagine.NewLocalStorage(imagine.LocalStorageParams{}),
        Cache:   imagine.NewMemoryStorage(imagine.MemoryStoreParams{}),
    })
    
    r.POST("/upload", gin.WrapF(img.UploadHandlerFunc()))
    r.GET("/images/*path", gin.WrapF(img.GetHandlerFunc()))
    
    r.Run(":8080")
}
```

#### Direct Image Processing

```go
// Process an image directly
params := &imagine.ImageParams{
    Width:   800,
    Height:  600,
    Fit:     "cover",
    Quality: 85,
    Format:  "webp",
}

processedImage, err := img.Get("image-hash.jpg", params)
if err != nil {
    log.Fatal(err)
}

// processedImage.Image contains the processed bytes
// processedImage.Type contains the MIME type
```

## 🎯 URL Parameters

Transform images by adding query parameters to the image URL:

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `w` | int | Width in pixels | `?w=800` |
| `h` | int | Height in pixels | `?h=600` |
| `fit` | string | Resize mode: `cover`, `contain`, `fill`, `inside`, `outside` | `?fit=cover` |
| `q`, `quality` | int | JPEG/WebP quality (1-100) | `?q=85` |
| `format` | string | Output format: `jpeg`, `png`, `webp`, `gif`, `tiff`, `avif` | `?format=webp` |
| `rotate` | int | Rotation angle: `0`, `90`, `180`, `270` | `?rotate=90` |
| `flip` | string | Flip direction: `h` (horizontal), `v` (vertical), `both` | `?flip=h` |
| `blur` | float | Gaussian blur (0.3-1000) | `?blur=5` |
| `sharpen` | float | Sharpen radius | `?sharpen=2` |
| `grayscale` | bool | Convert to grayscale | `?grayscale` |
| `gravity` | string | Crop position: `center`, `north`, `south`, `east`, `west`, `smart` | `?gravity=smart` |
| `thumbnail` | int | Square thumbnail size | `?thumbnail=150` |

### Example URLs

```
# Resize to 800x600 with smart cropping
/image.jpg?w=800&h=600&fit=cover&gravity=smart

# Create a 150x150 thumbnail
/image.jpg?thumbnail=150

# Convert to WebP with 85% quality
/image.jpg?format=webp&q=85

# Rotate 90 degrees and flip horizontally
/image.jpg?rotate=90&flip=h

# Apply blur effect and convert to grayscale
/image.jpg?blur=3&grayscale

# Complex transformation
/image.jpg?w=1920&h=1080&fit=cover&q=90&format=webp&sharpen=1&gravity=smart
```

## 💾 Storage Backends

Imagine supports multiple storage backends:

### Memory Storage (Development)
```go
storage := imagine.NewMemoryStorage(imagine.MemoryStoreParams{})
```

### Local Filesystem
```go
storage := imagine.NewLocalStorage(imagine.LocalStorageParams{
    Path: "/var/lib/imagine/images",
})
```

### Redis
```go
storage := imagine.NewRedisStorage(imagine.RedisStoreParams{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
    TTL:      24 * time.Hour,
})
```

### SQLite
```go
storage := imagine.NewSQLiteStorage(imagine.SQLiteStoreParams{
    Path: "/var/lib/imagine/images.db",
})
```

### BoltDB
```go
storage := imagine.NewBoltStorage(imagine.BoltStoreParams{
    Path:   "/var/lib/imagine/bolt.db",
    Bucket: "images",
})
```

## 🎭 Examples

### Upload an Image

```bash
curl -X POST -F "file=@photo.jpg" http://localhost:8080/upload
# Returns: abc123def456...
```

### Transform Images

```bash
# Resize to 800x600
curl http://localhost:8080/images/abc123def456.jpg?w=800&h=600

# Create a WebP thumbnail
curl http://localhost:8080/images/abc123def456.jpg?thumbnail=200&format=webp

# Apply multiple transformations
curl http://localhost:8080/images/abc123def456.jpg?w=1920&h=1080&fit=cover&q=90&rotate=180&flip=v
```

### Programmatic Upload

```go
// Upload an image programmatically
imageData, _ := ioutil.ReadFile("photo.jpg")
hash, err := img.Upload(imageData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Image uploaded: %s\n", hash)
```

## 📚 API Reference

### Core Types

```go
type Imagine struct {
    // Main application struct
}

type ImageParams struct {
    Width     int     // Target width
    Height    int     // Target height
    Quality   int     // JPEG/WebP quality (1-100)
    Format    string  // Output format
    Thumbnail int     // Thumbnail size
    Fit       string  // Resize mode
    Rotate    int     // Rotation angle
    Flip      string  // Flip direction
    Blur      float64 // Blur sigma
    Sharpen   float64 // Sharpen radius
    Grayscale bool    // Convert to grayscale
    Gravity   string  // Crop gravity
}

type ProcessedImage struct {
    Type  string // MIME type
    Image []byte // Image data
}
```

### Main Methods

```go
// Create a new Imagine instance
func New(params Params) (*Imagine, error)

// Process and retrieve an image
func (i *Imagine) Get(filename string, params *ImageParams) (*ProcessedImage, error)

// Upload a new image
func (i *Imagine) Upload(data []byte) (string, error)

// Parse URL parameters
func (i *Imagine) ParamsFromQueryString(query string) (*ImageParams, error)

// HTTP handlers
func (i *Imagine) UploadHandlerFunc() http.HandlerFunc
func (i *Imagine) GetHandlerFunc() http.HandlerFunc
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development

```bash
# Run tests
make test

# Run specific test
go test -run TestName ./...

# Build the project
go build ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [libvips](https://github.com/libvips/libvips) - The fantastic image processing library
- [bimg](https://github.com/h2non/bimg) - Go bindings for libvips
- All our [contributors](https://github.com/risico/imagine/graphs/contributors)

## 📞 Support

- 📫 [Open an issue](https://github.com/risico/imagine/issues)
- 💬 [Discussions](https://github.com/risico/imagine/discussions)
- 📖 [Documentation](https://pkg.go.dev/github.com/risico/imagine)

---

Made with ❤️ by the Imagine community