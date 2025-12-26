// news.go
package models

import (
	"federation-backend/app/db/models/enums"
	"time"
)

type BaseNewsData struct {
	Model
	Heading     string `json:"heading" gorm:"size:255"`
	Description string `json:"description" gorm:"type:text"`
	Images      []File `json:"images" gorm:"many2many:news_images"`
}

type News struct {
	BaseNewsData
	Date      time.Time `json:"date"`
	ChapterID uint      `json:"chapter_id"`
	Chapter   Chapter   `json:"chapter" gorm:"foreignKey:ChapterID"`
	Links     string    `json:"links"`
}

type HistoryItem struct {
	BaseNewsData
	Year int `json:"year"`
}

type Chapter struct {
	Model
	Name   string     `json:"name" gorm:"size:100"`
	BarIdx uint       `json:"bar_idx"`
	Page   enums.Page `json:"page"`
}
