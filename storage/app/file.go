package app

import "io"

type File struct {
	Name    string
	Size    int
	Type    string
	Content io.Reader
}

type FileDTO struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Resize string `json:"resize"`
}
