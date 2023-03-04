package server

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"github.com/juju/errors"
)

type ServerParams struct {
	Hostname string
	Port     int
}

type imageParams struct {
	Width, Height int
	Format        string
	Thumbnail     int
	Quality       int
}

func (p ServerParams) withDefaults() ServerParams {
	return p
}

type Server struct {
	params *ServerParams
}

func NewServer(params ServerParams) *Server {
	s := &Server{
		params: &params,
	}
	return s
}

func (s *Server) Start() error {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/upload", func(c *gin.Context) {
		// single file
		file, _ := c.FormFile("file")
		log.Println(file.Filename)

		uuid := uuid.New()
		filename := uuid.String()

		ct, ok := file.Header["Content-Type"]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not find content-type"})
			return
		}

		switch ct[0] {
		case "image/jpeg":
		case "image/jpg":
		case "image/png":
		case "image/gif":
		case "image/webp":
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "file type not supported"})
			return
		}

		// Upload the file to specific dst.
		tmpPath := fmt.Sprintf("uploads/temp/%s", filename)
		err := c.SaveUploadedFile(file, tmpPath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// open the file to generate its hash
		tmpFile, err := os.Open(tmpPath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer tmpFile.Close()

		hash := md5.New()
		_, err = io.Copy(hash, tmpFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := fmt.Sprintf("%x", hash.Sum(nil))

		newFilePath := fmt.Sprintf("uploads/%s", name)
		err = os.Rename(tmpPath, newFilePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"file": name})
	})

	r.GET("/:id/i.:format", func(c *gin.Context) {
		id := c.Param("id")

		params := imageParams{}

		wStr := c.Query("w")
		if wStr != "" {
			w, _ := strconv.Atoi(wStr)
			params.Width = w
		}

		hStr := c.Query("h")
		if hStr != "" {
			h, _ := strconv.Atoi(hStr)
			params.Height = h
		}

		qStr := c.Query("q")
		if hStr != "" {
			q, _ := strconv.Atoi(qStr)
			params.Quality = q
		}

		tStr := c.Query("t")
		if tStr != "" {
			t, _ := strconv.Atoi(tStr)
			params.Thumbnail = t
		}

		params.Format = c.Param("format")

		img, err := processImage(fmt.Sprintf("uploads/%s", id), params)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, fmt.Sprintf("image/%s", img.Type()), img.Image())
	})

	return errors.Trace(r.Run())
}

func processImage(path string, params imageParams) (*bimg.Image, error) {
	image, err := bimg.Read(path)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if params.Width > 0 && params.Height > 0 {
		image, err = bimg.NewImage(image).Resize(800, 600)
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
