package models

type File struct {
	Model
	Name string
	Size int64
	Path string `json:"path"`
}
