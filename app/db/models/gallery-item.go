package models

type GalleryItem struct {
	Images    []FilePath `json:"images" gorm:"many2many:gallery_item_images"`
	ChapterID uint
	Chapter   Chapter `json:"chapter" gorm:"foreignkey:ChapterID"`
}
