// gallery-item.go (no changes needed as it uses relations)
package models

type GalleryItem struct {
	Model
	Images    []File `json:"images" gorm:"many2many:gallery_item_images"`
	ChapterID uint
	Chapter   Chapter `json:"chapter" gorm:"foreignkey:ChapterID"`
}
