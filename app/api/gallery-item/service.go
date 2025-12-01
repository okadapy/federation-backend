package gallery_item

import (
	"errors"
	"federation-backend/app/db/models"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type CreateGalleryItemDTO struct {
	Name      string                  `form:"name"`
	ChapterID uint                    `form:"chapter_id" binding:"required"`
	Date      string                  `form:"date" binding:"required"`
	Images    []*multipart.FileHeader `form:"images" binding:"required,min=1"`
	Preview   *multipart.FileHeader   `form:"preview" binding:"required"`
}

type UpdateGalleryItemDTO struct {
	ChapterID *uint                   `form:"chapter_id"`
	Name      *string                 `form:"name"`
	Date      *string                 `form:"date"`
	Images    []*multipart.FileHeader `form:"images"`
	Preview   *multipart.FileHeader   `form:"preview" binding:"required"`
}
type Service struct {
	db          *gorm.DB
	fileService FileService
}

type FileService interface {
	SaveFile(fileHeader *multipart.FileHeader) (*models.File, error)
	DeleteFile(filename string) error
}

func (s *Service) parseDate(date string, dst *time.Time) error {
	timestamp, err := strconv.Atoi(date)
	if err != nil {
		return err
	}

	*dst = time.Unix(int64(timestamp), 0)
	return nil
}

func (s *Service) Create(dto interface{}) error {
	createDTO, ok := dto.(*CreateGalleryItemDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	var date time.Time
	if err := s.parseDate(createDTO.Date, &date); err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		preview, err := s.fileService.SaveFile(createDTO.Preview)
		if err != nil {
			return fmt.Errorf("failed to save image: %w", err)
		}

		galleryItem := models.GalleryItem{
			ChapterID: createDTO.ChapterID,
			Name:      createDTO.Name,
			Date:      date,
		}

		if err := tx.Create(&galleryItem).Error; err != nil {
			return fmt.Errorf("failed to create gallery item: %w", err)
		}

		if err := tx.Model(&galleryItem).Association("Preview").Replace(preview); err != nil {
			return fmt.Errorf("failed to add preview: %w", err)
		}

		// Save and associate images
		for _, fileHeader := range createDTO.Images {
			file, err := s.fileService.SaveFile(fileHeader)
			if err != nil {
				return fmt.Errorf("failed to save image: %w", err)
			}

			// Associate the file with gallery item using the many-to-many relationship
			if err := tx.Model(&galleryItem).Association("Images").Append(file); err != nil {
				return fmt.Errorf("failed to associate image: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) Get(id uint) (models.GalleryItem, error) {
	var item models.GalleryItem
	err := s.db.
		Preload("Preview").
		Preload("Images").
		Preload("Chapter").
		First(&item, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.GalleryItem{}, errors.New("gallery item not found")
		}
		return models.GalleryItem{}, fmt.Errorf("failed to get gallery item: %w", err)
	}
	return item, nil
}

func (s *Service) GetAll() ([]models.GalleryItem, error) {
	var items []models.GalleryItem
	err := s.db.
		Preload("Preview").
		Preload("Images").
		Preload("Chapter").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get gallery items: %w", err)
	}
	return items, nil
}

func (s *Service) Update(id uint, dto interface{}) error {
	updateDTO, ok := dto.(*UpdateGalleryItemDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var item models.GalleryItem
		if err := tx.Preload("Images").Preload("Preview").Preload("Chapter").First(&item, id).Error; err != nil {
			return fmt.Errorf("gallery item not found: %w", err)
		}

		if updateDTO.Preview != nil {
			file, err := s.fileService.SaveFile(updateDTO.Preview)
			if err != nil {
				return fmt.Errorf("failed to update gallery item: %w", err)
			}
			tx.Model(&item).Association("Preview").Replace(file)
		}

		if updateDTO.ChapterID != nil {
			item.ChapterID = *updateDTO.ChapterID
		}
		if updateDTO.Name != nil {
			item.Name = *updateDTO.Name
		}
		if updateDTO.Date != nil {
			var date time.Time
			if err := s.parseDate(*updateDTO.Date, &date); err != nil {
				return err
			}
			item.Date = date
		}

		// Handle image updates if new images are provided
		if updateDTO.Images != nil && len(updateDTO.Images) > 0 {
			// Get current images to delete them later
			currentImages := make([]models.File, len(item.Images))
			copy(currentImages, item.Images)

			// Clear existing images from association
			if err := tx.Model(&item).Association("Images").Clear(); err != nil {
				return fmt.Errorf("failed to clear existing images: %w", err)
			}

			// Delete the old file records and physical files
			for _, image := range currentImages {
				// Extract just the filename from the path for deletion
				filename := filepath.Base(image.Path)
				if err := s.fileService.DeleteFile(filename); err != nil {
					// Log the error but continue with other deletions
					fmt.Printf("Warning: failed to delete image file %s: %v\n", filename, err)
				}

				// Delete the file record from database
				if err := tx.Delete(&image).Error; err != nil {
					fmt.Printf("Warning: failed to delete file record %d: %v\n", image.Id, err)
				}
			}

			// Add new images
			for _, fileHeader := range updateDTO.Images {
				file, err := s.fileService.SaveFile(fileHeader)
				if err != nil {
					return fmt.Errorf("failed to save image: %w", err)
				}

				if err := tx.Model(&item).Association("Images").Append(file); err != nil {
					return fmt.Errorf("failed to associate image: %w", err)
				}
			}
		}

		// Save the updated gallery item
		if err := tx.Save(&item).Error; err != nil {
			return fmt.Errorf("failed to update gallery item: %w", err)
		}

		return nil
	})
}

func (s *Service) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var item models.GalleryItem
		if err := tx.Preload("Images").Preload("Preview").First(&item, id).Error; err != nil {
			return fmt.Errorf("gallery item not found: %w", err)
		}

		filename := filepath.Base(item.Preview.Path)
		if err := s.fileService.DeleteFile(filename); err != nil {
			fmt.Printf("Warning: failed to delete image file %s: %v\n", filename, err)
		}

		// Delete associated files
		for _, image := range item.Images {
			// Extract just the filename from the path
			filename := filepath.Base(image.Path)
			if err := s.fileService.DeleteFile(filename); err != nil {
				// Log but continue with other deletions
				fmt.Printf("Warning: failed to delete image file %s: %v\n", filename, err)
			}

			// Delete the file record from database
			if err := tx.Delete(&image).Error; err != nil {
				fmt.Printf("Warning: failed to delete file record %d: %v\n", image.Id, err)
			}
		}

		// Clear the association first (good practice)
		if err := tx.Model(&item).Association("Images").Clear(); err != nil {
			return fmt.Errorf("failed to clear image associations: %w", err)
		}

		// Delete the gallery item
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
