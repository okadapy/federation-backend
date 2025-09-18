package gallery_item

import (
	"errors"
	"federation-backend/app/db/models"
	"fmt"
	"mime/multipart"

	"gorm.io/gorm"
)

type CreateGalleryItemDTO struct {
	ChapterID uint                    `form:"chapterId" binding:"required"`
	Images    []*multipart.FileHeader `form:"images" binding:"required,min=1"`
}

type UpdateGalleryItemDTO struct {
	ChapterID *uint                   `form:"chapterId"`
	Images    []*multipart.FileHeader `form:"images"`
}

type Service struct {
	db          *gorm.DB
	fileService FileService
}

type FileService interface {
	SaveFile(fileHeader *multipart.FileHeader) (string, error)
	DeleteFile(filename string) error
}

func (s *Service) Create(dto interface{}) error {
	createDTO, ok := dto.(*CreateGalleryItemDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		galleryItem := models.GalleryItem{
			ChapterID: createDTO.ChapterID,
		}

		if err := tx.Create(&galleryItem).Error; err != nil {
			return fmt.Errorf("failed to create gallery item: %w", err)
		}

		for _, fileHeader := range createDTO.Images {
			filename, err := s.fileService.SaveFile(fileHeader)
			if err != nil {
				return fmt.Errorf("failed to save image: %w", err)
			}

			file := models.File{
				Name: fileHeader.Filename,
				Size: fileHeader.Size,
				Path: filename,
			}
			if err := tx.Create(&file).Error; err != nil {
				return fmt.Errorf("failed to create file record: %w", err)
			}

			if err := tx.Model(&galleryItem).Association("Images").Append(&file); err != nil {
				return fmt.Errorf("failed to associate image: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) Get(id uint) (models.GalleryItem, error) {
	var item models.GalleryItem
	err := s.db.Preload("Images").Preload("Chapter").First(&item, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.GalleryItem{}, errors.New("gallery item not found")
		}
		return models.GalleryItem{}, fmt.Errorf("failed to get gallery item: %w", err)
	}
	return item, nil
}

func (s *Service) Update(id uint, dto interface{}) error {
	updateDTO, ok := dto.(*UpdateGalleryItemDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var item models.GalleryItem
		if err := tx.First(&item, id).Error; err != nil {
			return fmt.Errorf("gallery item not found: %w", err)
		}

		if updateDTO.ChapterID != nil {
			item.ChapterID = *updateDTO.ChapterID
		}

		if updateDTO.Images != nil {
			// Clear existing images
			if err := tx.Model(&item).Association("Images").Clear(); err != nil {
				return fmt.Errorf("failed to clear existing images: %w", err)
			}

			// Add new images
			for _, fileHeader := range updateDTO.Images {
				filename, err := s.fileService.SaveFile(fileHeader)
				if err != nil {
					return fmt.Errorf("failed to save image: %w", err)
				}

				file := models.File{
					Name: fileHeader.Filename,
					Size: fileHeader.Size,
					Path: filename,
				}
				if err := tx.Create(&file).Error; err != nil {
					return fmt.Errorf("failed to create file record: %w", err)
				}

				if err := tx.Model(&item).Association("Images").Append(&file); err != nil {
					return fmt.Errorf("failed to associate image: %w", err)
				}
			}
		}

		if err := tx.Save(&item).Error; err != nil {
			return fmt.Errorf("failed to update gallery item: %w", err)
		}

		return nil
	})
}

func (s *Service) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var item models.GalleryItem
		if err := tx.Preload("Images").First(&item, id).Error; err != nil {
			return fmt.Errorf("gallery item not found: %w", err)
		}

		// Delete associated files
		for _, image := range item.Images {
			if err := s.fileService.DeleteFile(image.Path); err != nil {
				return fmt.Errorf("failed to delete image file: %w", err)
			}
		}

		if err := tx.Delete(&item).Error; err != nil {
			return fmt.Errorf("failed to delete gallery item: %w", err)
		}

		return nil
	})
}

func NewService(db *gorm.DB, fileService FileService) *Service {
	return &Service{
		db:          db,
		fileService: fileService,
	}
}
