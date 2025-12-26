package news

import (
	"errors"
	"federation-backend/app/api/shared"
	"federation-backend/app/db/models"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type CreateNewsDTO struct {
	Heading     string                  `form:"heading" binding:"required"`
	Description string                  `form:"description" binding:"required"`
	Date        string                  `form:"date" binding:"required"`
	ChapterID   uint                    `form:"chapterId" binding:"required"`
	Links       *string                 `form:"links"`
	Images      []*multipart.FileHeader `form:"images" binding:"required,min=1"`
}

type UpdateNewsDTO struct {
	Heading       *string                 `form:"heading"`
	Description   *string                 `form:"description"`
	Date          *string                 `form:"date"`
	ChapterID     *uint                   `form:"chapterId"`
	Links         *string                 `form:"links"`
	NewImages     []*multipart.FileHeader `form:"newImages"`
	DeletedImages []uint                  `form:"deletedImages"`
}

type Service struct {
	db          *gorm.DB
	fileService shared.FileProcessor
}

type FileService interface {
	SaveFile(fileHeader *multipart.FileHeader) (*models.File, error)
	DeleteFile(filename string) error
}

func (s *Service) parseDate(date string) (time.Time, error) {
	// Try parsing as timestamp first
	if timestamp, err := strconv.ParseInt(date, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}

	// Try parsing as RFC3339 or other date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, date); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", date)
}

func (s *Service) Create(dto interface{}) error {
	createDTO, ok := dto.(*CreateNewsDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	date, err := s.parseDate(createDTO.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		news := models.News{
			BaseNewsData: models.BaseNewsData{
				Heading:     createDTO.Heading,
				Description: createDTO.Description,
			},
			Date:      date,
			ChapterID: createDTO.ChapterID,
			Links:     *createDTO.Links,
		}

		if err := tx.Create(&news).Error; err != nil {
			return fmt.Errorf("failed to create news: %w", err)
		}

		// Save and associate images
		for _, fileHeader := range createDTO.Images {
			file, err := s.fileService.SaveFile(fileHeader)
			if err != nil {
				return fmt.Errorf("failed to save image: %w", err)
			}

			// Use GORM's association method
			if err := tx.Model(&news).Association("Images").Append(file); err != nil {
				return fmt.Errorf("failed to associate image: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) Get(id uint) (models.News, error) {
	var news models.News
	err := s.db.
		Preload("Images").
		Preload("Chapter").
		First(&news, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.News{}, errors.New("news not found")
		}
		return models.News{}, fmt.Errorf("failed to get news: %w", err)
	}
	return news, nil
}
func (s *Service) Update(id uint, dto interface{}) error {
	updateDTO, ok := dto.(*UpdateNewsDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var news models.News
		if err := tx.Preload("Images").First(&news, id).Error; err != nil {
			return fmt.Errorf("news not found: %w", err)
		}

		// Update basic fields
		if updateDTO.Heading != nil {
			news.Heading = *updateDTO.Heading
		}
		if updateDTO.Description != nil {
			news.Description = *updateDTO.Description
		}
		if updateDTO.Date != nil {
			date, err := s.parseDate(*updateDTO.Date)
			if err != nil {
				return fmt.Errorf("failed to parse date: %w", err)
			}
			news.Date = date
		}
		if updateDTO.Links != nil {
			news.Links = *updateDTO.Links
		}

		if updateDTO.ChapterID != nil {
			news.ChapterID = *updateDTO.ChapterID

			// Загружаем новый Chapter для обновления ассоциации
			var chapter models.Chapter
			if err := tx.First(&chapter, *updateDTO.ChapterID).Error; err != nil {
				return fmt.Errorf("failed to find chapter: %w", err)
			}

			// Обновляем ассоциацию
			if err := tx.Model(&news).Association("Chapter").Replace(&chapter); err != nil {
				return fmt.Errorf("failed to update chapter association: %w", err)
			}
		}

		// Handle image deletion
		if err := s.deleteImages(tx, &news, updateDTO.DeletedImages); err != nil {
			return err
		}

		// Handle new image addition (parallel)
		if len(updateDTO.NewImages) > 0 {
			if err := s.addNewImages(tx, &news, updateDTO.NewImages); err != nil {
				return err
			}
		}

		// Save the updated news item
		if err := tx.Save(&news).Error; err != nil {
			return fmt.Errorf("failed to update news: %w", err)
		}

		return nil
	})
}

// deleteImages удаляет указанные изображения
func (s *Service) deleteImages(tx *gorm.DB, news *models.News, deletedImageIDs []uint) error {
	if len(deletedImageIDs) == 0 {
		return nil
	}

	// Находим изображения для удаления
	var imagesToDelete []models.File
	if err := tx.Where("id IN ?", deletedImageIDs).Find(&imagesToDelete).Error; err != nil {
		return fmt.Errorf("failed to find images to delete: %w", err)
	}

	// Удаляем каждое изображение
	for _, image := range imagesToDelete {
		if err := s.deleteSingleImage(tx, news, &image); err != nil {
			// Логируем ошибку, но продолжаем удаление остальных
			fmt.Printf("Warning: %v\n", err)
		}
	}

	return nil
}

// deleteSingleImage удаляет одно изображение
func (s *Service) deleteSingleImage(tx *gorm.DB, news *models.News, image *models.File) error {
	// Удаляем из файловой системы
	filename := filepath.Base(image.Path)
	if err := s.fileService.DeleteFile(filename); err != nil {
		return fmt.Errorf("failed to delete image file %s: %w", filename, err)
	}

	// Удаляем из ассоциации
	if err := tx.Model(news).Association("Images").Delete(image); err != nil {
		return fmt.Errorf("failed to remove image %d from association: %w", image.Id, err)
	}

	// Удаляем запись из базы
	if err := tx.Delete(image).Error; err != nil {
		return fmt.Errorf("failed to delete file record %d: %w", image.Id, err)
	}

	return nil
}

// addNewImages добавляет новые изображения параллельно
func (s *Service) addNewImages(tx *gorm.DB, news *models.News, newImages []*multipart.FileHeader) error {
	// Сохраняем файлы параллельно
	files, errors := s.fileService.SaveFilesParallel(newImages)

	// Проверяем ошибки
	var saveErrors []error
	for i, err := range errors {
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("image %d: %w", i, err))
		}
	}

	if len(saveErrors) > 0 {
		// Удаляем успешно сохраненные файлы при наличии ошибок
		for i, file := range files {
			if file != nil && errors[i] == nil {
				filename := filepath.Base(file.Path)
				s.fileService.DeleteFile(filename)
			}
		}
		return fmt.Errorf("failed to save some images: %v", saveErrors)
	}

	// Ассоциируем успешно сохраненные файлы
	for _, file := range files {
		if file != nil {
			if err := tx.Model(news).Association("Images").Append(file); err != nil {
				return fmt.Errorf("failed to associate image: %w", err)
			}
		}
	}

	return nil
}

func (s *Service) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var news models.News
		if err := tx.Preload("Images").First(&news, id).Error; err != nil {
			return fmt.Errorf("news not found: %w", err)
		}

		// Delete associated files
		for _, image := range news.Images {
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
		if err := tx.Model(&news).Association("Images").Clear(); err != nil {
			return fmt.Errorf("failed to clear image associations: %w", err)
		}

		// Delete the news item
		if err := tx.Delete(&news).Error; err != nil {
			return fmt.Errorf("failed to delete news: %w", err)
		}

		return nil
	})
}

func (s *Service) GetAll() ([]models.News, error) {
	var news []models.News
	err := s.db.
		Preload("Images").
		Preload("Chapter").
		Find(&news).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}
	return news, nil
}

func NewService(db *gorm.DB, fileProcessor shared.FileProcessor) *Service {
	return &Service{
		db:          db,
		fileService: fileProcessor,
	}
}
