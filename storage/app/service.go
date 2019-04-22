package app

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/spf13/afero"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

var Service *service

func init() {
	if len(os.Getenv("TESTING")) > 0 {
		Service = NewService(afero.NewMemMapFs())
	} else {
		Service = NewService(afero.NewOsFs())
	}
}

type service struct {
	fs afero.Fs
}

func NewService(fs afero.Fs) *service {
	return &service{
		fs: fs,
	}
}

func (s service) SaveFile(file File) (string, error) {
	if !checkMimeType(file.Type) {
		return "", errors.New("wrong mime type")
	}

	path := getSavePath(file.Name)

	to, err := s.fs.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(to, file.Content)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s service) Resize(file File) (string, error) {
	img, _, err := image.Decode(file.Content)
	if err != nil {
		return "", err
	}
	thumbnail := resize.Thumbnail(100, 100, img, resize.Lanczos3)
	var buff bytes.Buffer
	switch file.Type {
	case "image/png":
		png.Encode(&buff, thumbnail)
		break
	case "image/jpg", "image/jpeg":
		jpeg.Encode(&buff, thumbnail, nil)
		break
	default:
		return "", errors.New("could not encode image")
	}
	if buff.Len() > 0 {
		path, err := s.SaveFile(File{
			Name:    fmt.Sprintf("thumb_%s", file.Name),
			Type:    file.Type,
			Content: bytes.NewReader(buff.Bytes()),
			Size:    buff.Len(),
		})
		if err != nil {
			return "", err
		}
		return path, nil
	}
	return "", errors.New("could not resize image")
}

func getSavePath(name string) string {
	return fmt.Sprintf("/images/%s", name)
}

func checkMimeType(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return true
	case "image/png":
		return true
	default:
		return false
	}
}
