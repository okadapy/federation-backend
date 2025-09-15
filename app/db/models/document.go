package models

import "federation-backend/app/db/models/enums"

type Document struct {
	Model
	FilePath
	Name    string
	Chapter enums.Doctype `json:"chapter" gorm:"type:enum('rules', 'regulations')"`
}
