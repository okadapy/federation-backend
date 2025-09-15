package models

import "federation-backend/app/db/models/enums"

type BaseNewsData struct {
	Model
	Heading     string     `json:"heading"`
	Description string     `json:"description"`
	Images      []FilePath `json:"images" gorm:"many2many:news_images"`
}

type News struct {
	BaseNewsData
	Date      string  `json:"date"`
	ChapterID uint    `json:"chapter_id"`
	Chapter   Chapter `json:"chapter" gorm:"foreignKey:ChapterID"`
}

type HistoryItem struct {
	BaseNewsData
	Year int `json:"year"`
}

type Chapter struct {
	Model
	Name string     `json:"name"`
	Page enums.Page `json:"page" gorm:"type:enum('news', 'gallery', 'documents')"`
}
