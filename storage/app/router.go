package app

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"html"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	setupRouter(router)
	return router
}

func setupRouter(router *gin.Engine) {
	api := router.Group("/storage")

	api.GET("/ping", ping)
	api.POST("/upload", upload)
	api.POST("/upload/link", link)
	api.POST("/upload/json", json)
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		errorResponse(c, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	files := form.File["images[]"]
	var paths []FileDTO
	for _, file := range files {
		reader, err := file.Open()
		if err != nil {
			errorResponse(c, fmt.Sprintf("could not open file to read: %s", err.Error()))
			return
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			errorResponse(c, fmt.Sprintf("could not read file: %s", err.Error()))
			return
		}
		name := html.UnescapeString(file.Filename)
		img, err := Service.SaveFile(File{
			Name:    name,
			Size:    int(file.Size),
			Type:    http.DetectContentType(b),
			Content: bytes.NewReader(b),
		})
		if err != nil {
			errorResponse(c, fmt.Sprintf("could not save file: %s", err.Error()))
			return
		}
		resize, err := Service.Resize(File{
			Name:    name,
			Size:    int(file.Size),
			Type:    http.DetectContentType(b),
			Content: bytes.NewReader(b),
		})
		if err != nil {
			errorResponse(c, fmt.Sprintf("could not resize file: %s", err.Error()))
			return
		}

		paths = append(paths, FileDTO{
			Name:   name,
			Path:   img,
			Resize: resize,
		})
	}

	// success
	c.JSON(http.StatusOK, paths)
}

func link(c *gin.Context) {
	url := c.PostForm("url")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	file, err := client.Get(url)
	if err != nil {
		errorResponse(c, fmt.Sprintf("error downloading file: %s", err.Error()))
		return
	}
	defer file.Body.Close()

	content, err := ioutil.ReadAll(file.Body)
	var paths []FileDTO
	img, err := Service.SaveFile(File{
		Name:    path.Base(url),
		Size:    len(content),
		Type:    http.DetectContentType(content),
		Content: bytes.NewReader(content),
	})
	if err != nil {
		errorResponse(c, fmt.Sprintf("could not save file: %s", err.Error()))
		return
	}
	resize, err := Service.Resize(File{
		Name:    path.Base(url),
		Size:    len(content),
		Type:    http.DetectContentType(content),
		Content: bytes.NewReader(content),
	})
	paths = append(paths, FileDTO{
		Name:   path.Base(url),
		Path:   img,
		Resize: resize,
	})
	// success
	c.JSON(http.StatusOK, paths)
}

func json(c *gin.Context) {
	data := new([]struct {
		Name    string `json:"name" binding:"required"`
		Size    int    `json:"size" binding:"required"`
		Type    string `json:"type" binding:"required"`
		Content string `json:"content" binding:"required"`
	})
	err := c.BindJSON(data)
	if err != nil {
		errorResponse(c, fmt.Sprintf("could not unmarshal files: %s", err.Error()))
		return
	}

	var paths []FileDTO
	for _, file := range *data {
		// Get base64 value
		b64data := file.Content[strings.IndexByte(file.Content, ',')+1:]
		data, err := base64.StdEncoding.DecodeString(b64data)
		if err != nil {
			errorResponse(c, fmt.Sprintf("could not decode base64 file string: %s", err.Error()))
			return
		}
		name := html.UnescapeString(file.Name)
		img, err := Service.SaveFile(File{
			Name:    name,
			Size:    int(file.Size),
			Type:    file.Type,
			Content: bytes.NewReader(data),
		})
		if err != nil {
			errorResponse(c, fmt.Sprintf("could not save file: %s", err.Error()))
			return
		}
		resize, err := Service.Resize(File{
			Name:    name,
			Size:    int(file.Size),
			Type:    file.Type,
			Content: bytes.NewReader(data),
		})
		if err != nil {
			errorResponse(c, fmt.Sprintf("could not resize file: %s", err.Error()))
			return
		}

		paths = append(paths, FileDTO{
			Name:   name,
			Path:   img,
			Resize: resize,
		})

	}
	c.JSON(http.StatusOK, paths)
}

func errorResponse(c *gin.Context, mess string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": mess,
	})
}
