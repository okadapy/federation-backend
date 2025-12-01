// gallery-item.go (no changes needed as it uses relations)
package models

import "time"

type GalleryItem struct {
	Model
	Name      string    `json:"name"`
	Date      time.Time `json:"date"`
	PreviewID uint
	Preview   File   `json:"preview" gorm:"foreignKey:PreviewID"`
	Images    []File `json:"images" gorm:"many2many:gallery_item_images"`
	ChapterID uint
	Chapter   Chapter `json:"chapter" gorm:"foreignkey:ChapterID"`
}
