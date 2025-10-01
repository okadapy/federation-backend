package chapter

import (
	"errors"
	"federation-backend/app/db/models"
	"federation-backend/app/db/models/enums"
	"fmt"

	"gorm.io/gorm"
)

type CreateChapterDTO struct {
	Name string     `form:"name" binding:"required"`
	Page enums.Page `form:"page" binding:"required"`
}

type UpdateChapterDTO struct {
	Name *string     `form:"name"`
	Page *enums.Page `form:"page"`
}

type Service struct {
	db *gorm.DB
}

func (s *Service) Create(dto interface{}) error {
	createDTO, ok := dto.(*CreateChapterDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	chapter := models.Chapter{
		Name: createDTO.Name,
		Page: createDTO.Page,
	}

	if err := s.db.Create(&chapter).Error; err != nil {
		return fmt.Errorf("failed to create chapter: %w", err)
	}

	return nil
}

func (s *Service) Get(id uint) (models.Chapter, error) {
	var chapter models.Chapter
	err := s.db.First(&chapter, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Chapter{}, errors.New("chapter not found")
		}
		return models.Chapter{}, fmt.Errorf("failed to get chapter: %w", err)
	}
	return chapter, nil
}

func (s *Service) Update(id uint, dto interface{}) error {
	updateDTO, ok := dto.(*UpdateChapterDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	var chapter models.Chapter
	if err := s.db.First(&chapter, id).Error; err != nil {
		return fmt.Errorf("chapter not found: %w", err)
	}

	if updateDTO.Name != nil {
		chapter.Name = *updateDTO.Name
	}
	if updateDTO.Page != nil {
		chapter.Page = *updateDTO.Page
	}

	if err := s.db.Save(&chapter).Error; err != nil {
		return fmt.Errorf("failed to update chapter: %w", err)
	}

	return nil
}

func (s *Service) Delete(id uint) error {
	var chapter models.Chapter
	if err := s.db.First(&chapter, id).Error; err != nil {
		return fmt.Errorf("chapter not found: %w", err)
	}

	if err := s.db.Delete(&chapter).Error; err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}

	return nil
}

func (s *Service) GetAll() ([]models.Chapter, error) {
	var chapters []models.Chapter
	if err := s.db.Find(&chapters).Error; err != nil {
		return nil, fmt.Errorf("failed to get chapters: %w", err)
	}
	return chapters, nil
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}
