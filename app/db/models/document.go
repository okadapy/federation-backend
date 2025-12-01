// document.go
package models

import "federation-backend/app/db/models/enums"

type Document struct {
	Model
	File
	Name    string        `gorm:"size:255"`
	Chapter enums.Doctype `json:"chapter"`
}
