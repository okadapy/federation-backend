// file.go
package models

type File struct {
	Model
	Name string `json:"name" gorm:"size:255"`
	Size int64  `json:"size"`
	Path string `json:"path" gorm:"size:500"`
}
