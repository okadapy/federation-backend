package news

import (
	"errors"
	"federation-backend/app/db/models"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type CreateNewsDTO struct {
	Heading     string                  `form:"heading" binding:"required"`
	Description string                  `form:"description" binding:"required"`
	Date        string                  `form:"date" binding:"required"`
	ChapterID   uint                    `form:"chapterId" binding:"required"`
	Images      []*multipart.FileHeader `form:"images" binding:"required,min=1"`
}

type UpdateNewsDTO struct {
	Heading     *string                 `form:"heading"`
	Description *string                 `form:"description"`
	Date        *string                 `form:"date"`
	ChapterID   *uint                   `form:"chapterId"`
	Images      []*multipart.FileHeader `form:"images"`
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
	createDTO, ok := dto.(*CreateNewsDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	var date time.Time
	if err := s.parseDate(createDTO.Date, &date); err != nil {
		return err
	}

	news := models.News{
		BaseNewsData: models.BaseNewsData{
			Heading:     createDTO.Heading,
			Description: createDTO.Description,
			Images:      nil,
		},
		Date:      date,
		ChapterID: createDTO.ChapterID,
	}
	if err := s.db.Create(&news).Error; err != nil {
		return fmt.Errorf("failed to create news: %w", err)
	}

	for _, fileHeader := range createDTO.Images {
		file, err := s.fileService.SaveFile(fileHeader)
		if err != nil {
			return fmt.Errorf("failed to save image: %w", err)
		}
		if err := s.db.Model(&news).Association("Images").Append(&file); err != nil {
			return fmt.Errorf("failed to associate image: %w", err)
		}
	}

	return nil
}

func (s *Service) Get(id uint) (models.News, error) {
	var news models.News
	err := s.db.Preload("Images").Preload("Chapter").First(&news, id).Error
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

	var news models.News
	if err := s.db.First(&news, id).Error; err != nil {
		return fmt.Errorf("news not found: %w", err)
	}

	if updateDTO.Heading != nil {
		news.Heading = *updateDTO.Heading
	}
	if updateDTO.Description != nil {
		news.Description = *updateDTO.Description
	}
	if updateDTO.Date != nil {
		var date time.Time
		if err := s.parseDate(*updateDTO.Date, &date); err != nil {
			return err
		}
		news.Date = date
	}
	if updateDTO.ChapterID != nil {
		news.ChapterID = *updateDTO.ChapterID
	}

	if updateDTO.Images != nil {
		var existingImages []models.File
		if err := s.db.Model(&news).Association("Images").Find(&existingImages); err != nil {
			return fmt.Errorf("failed to find existing images: %w", err)
		}
		for _, image := range existingImages {
			if err := s.fileService.DeleteFile(image.Path); err != nil {
				return fmt.Errorf("failed to delete image file: %w", err)
			}
		}
		if err := s.db.Model(&news).Association("Images").Clear(); err != nil {
			return fmt.Errorf("failed to clear existing images: %w", err)
		}

		// Add new images
		for _, fileHeader := range updateDTO.Images {
			file, err := s.fileService.SaveFile(fileHeader)
			if err != nil {
				return fmt.Errorf("failed to save image: %w", err)
			}
			if err := s.db.Model(&news).Association("Images").Append(&file); err != nil {
				return fmt.Errorf("failed to associate image: %w", err)
			}
		}
	}

	if err := s.db.Save(&news).Error; err != nil {
		return fmt.Errorf("failed to update news: %w", err)
	}

	return nil
}

func (s *Service) Delete(id uint) error {
	var news models.News
	if err := s.db.Preload("Images").First(&news, id).Error; err != nil {
		return fmt.Errorf("news not found: %w", err)
	}

	// Delete associated files
	for _, image := range news.Images {
		if err := s.fileService.DeleteFile(image.Path); err != nil {
			continue
		}
	}

	if err := s.db.Delete(&news).Error; err != nil {
		return fmt.Errorf("failed to delete news: %w", err)
	}

	return nil
}

func (s *Service) GetAll() ([]models.News, error) {
	var news []models.News
	if err := s.db.Preload("Images").Find(&news).Error; err != nil {
		return nil, fmt.Errorf("failed to find news: %w", err)
	}

	return news, nil
}

func NewService(db *gorm.DB, fileService FileService) *Service {
	return &Service{
		db:          db,
		fileService: fileService,
	}
}
