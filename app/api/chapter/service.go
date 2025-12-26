package chapter

import (
	"errors"
	"federation-backend/app/db/models"
	"federation-backend/app/db/models/enums"
	"fmt"

	"gorm.io/gorm"
)

type CreateChapterDTO struct {
	Name   string     `form:"name" binding:"required"`
	Page   enums.Page `form:"page" binding:"required"`
	BarIdx *uint      `form:"bar_idx,omitempty"`
}

type UpdateChapterDTO struct {
	Name   *string     `form:"name"`
	BarIdx *uint       `form:"bar_idx"`
	Page   *enums.Page `form:"page"`
}

type Service struct {
	db *gorm.DB
}

func (s *Service) Create(dto interface{}) error {
	createDTO, ok := dto.(*CreateChapterDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Определяем порядок для новой главы
		var barIdx uint = 1

		// Если указан явный индекс, проверяем и освобождаем место
		if createDTO.BarIdx != nil {
			barIdx = *createDTO.BarIdx

			// Освобождаем место: сдвигаем существующие элементы вниз
			if err := tx.Model(&models.Chapter{}).
				Where("page = ? AND bar_idx >= ?", createDTO.Page, barIdx).
				Update("bar_idx", gorm.Expr("bar_idx + 1")).Error; err != nil {
				return fmt.Errorf("failed to reorder chapters: %w", err)
			}
		} else {
			// Если индекс не указан, ставим в конец
			var maxIdx uint
			err := tx.Model(&models.Chapter{}).
				Where("page = ?", createDTO.Page).
				Select("COALESCE(MAX(bar_idx), 0)").
				Scan(&maxIdx).Error
			if err != nil {
				return fmt.Errorf("failed to get max bar_idx: %w", err)
			}
			barIdx = maxIdx + 1
		}

		// Создаем главу
		chapter := models.Chapter{
			Name:   createDTO.Name,
			Page:   createDTO.Page,
			BarIdx: barIdx,
		}

		if err := tx.Create(&chapter).Error; err != nil {
			return fmt.Errorf("failed to create chapter: %w", err)
		}

		return nil
	})
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

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Получаем текущую главу
		var chapter models.Chapter
		if err := tx.First(&chapter, id).Error; err != nil {
			return fmt.Errorf("chapter not found: %w", err)
		}

		oldBarIdx := chapter.BarIdx
		oldPage := chapter.Page
		newBarIdx := updateDTO.BarIdx
		newPage := updateDTO.Page

		// Обновляем поля, если они переданы
		if updateDTO.Name != nil {
			chapter.Name = *updateDTO.Name
		}

		// Обрабатываем изменение порядка и/или страницы
		if newBarIdx != nil || newPage != nil {
			// Определяем новые значения
			targetBarIdx := oldBarIdx
			targetPage := oldPage

			if newBarIdx != nil {
				targetBarIdx = *newBarIdx
			}
			if newPage != nil {
				targetPage = *newPage
			}

			// Если страница изменилась
			if newPage != nil && *newPage != oldPage {
				// 1. Удаляем главу со старой страницы (сдвигаем остальные вверх)
				if err := tx.Model(&models.Chapter{}).
					Where("page = ? AND bar_idx > ?", oldPage, oldBarIdx).
					Update("bar_idx", gorm.Expr("bar_idx - 1")).Error; err != nil {
					return fmt.Errorf("failed to reorder old page: %w", err)
				}

				// 2. Освобождаем место на новой странице (сдвигаем вниз)
				if err := tx.Model(&models.Chapter{}).
					Where("page = ? AND bar_idx >= ?", targetPage, targetBarIdx).
					Update("bar_idx", gorm.Expr("bar_idx + 1")).Error; err != nil {
					return fmt.Errorf("failed to reorder new page: %w", err)
				}

				chapter.Page = targetPage
				chapter.BarIdx = targetBarIdx
			} else if newBarIdx != nil && *newBarIdx != oldBarIdx {
				// Только индекс изменился, страница та же
				if *newBarIdx < oldBarIdx {
					// Сдвигаем вниз элементы между новым и старым индексом
					if err := tx.Model(&models.Chapter{}).
						Where("page = ? AND bar_idx >= ? AND bar_idx < ?", oldPage, *newBarIdx, oldBarIdx).
						Update("bar_idx", gorm.Expr("bar_idx + 1")).Error; err != nil {
						return fmt.Errorf("failed to reorder chapters up: %w", err)
					}
				} else {
					// Сдвигаем вверх элементы между старым и новым индексом
					if err := tx.Model(&models.Chapter{}).
						Where("page = ? AND bar_idx > ? AND bar_idx <= ?", oldPage, oldBarIdx, *newBarIdx).
						Update("bar_idx", gorm.Expr("bar_idx - 1")).Error; err != nil {
						return fmt.Errorf("failed to reorder chapters down: %w", err)
					}
				}
				chapter.BarIdx = *newBarIdx
			}
		}

		// Сохраняем изменения
		if err := tx.Save(&chapter).Error; err != nil {
			return fmt.Errorf("failed to update chapter: %w", err)
		}

		return nil
	})
}

func (s *Service) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Получаем удаляемую главу
		var chapter models.Chapter
		if err := tx.First(&chapter, id).Error; err != nil {
			return fmt.Errorf("chapter not found: %w", err)
		}

		// Удаляем главу
		if err := tx.Delete(&chapter).Error; err != nil {
			return fmt.Errorf("failed to delete chapter: %w", err)
		}

		// Сдвигаем остальные главы на странице вверх
		if err := tx.Model(&models.Chapter{}).
			Where("page = ? AND bar_idx > ?", chapter.Page, chapter.BarIdx).
			Update("bar_idx", gorm.Expr("bar_idx - 1")).Error; err != nil {
			return fmt.Errorf("failed to reorder after delete: %w", err)
		}

		return nil
	})
}

func (s *Service) GetAll() ([]models.Chapter, error) {
	var chapters []models.Chapter
	if err := s.db.Find(&chapters).Error; err != nil {
		return nil, fmt.Errorf("failed to get chapters: %w", err)
	}
	return chapters, nil
}

// GetByPage возвращает главы для конкретной страницы в правильном порядке
func (s *Service) GetByPage(page enums.Page) ([]models.Chapter, error) {
	var chapters []models.Chapter
	if err := s.db.
		Where("page = ?", page).
		Order("bar_idx ASC").
		Find(&chapters).Error; err != nil {
		return nil, fmt.Errorf("failed to get chapters by page: %w", err)
	}
	return chapters, nil
}

// Reorder полностью пересортирует порядок глав на странице
func (s *Service) Reorder(page enums.Page, chapterIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Получаем все главы на странице
		var chapters []models.Chapter
		if err := tx.Where("page = ?", page).Find(&chapters).Error; err != nil {
			return fmt.Errorf("failed to get chapters: %w", err)
		}

		// Создаем мапу для быстрого доступа
		chapterMap := make(map[uint]*models.Chapter)
		for i := range chapters {
			chapterMap[chapters[i].Id] = &chapters[i]
		}

		// Обновляем индексы согласно новому порядку
		for newIdx, chapterID := range chapterIDs {
			if chapter, exists := chapterMap[chapterID]; exists {
				chapter.BarIdx = uint(newIdx + 1)
				if err := tx.Save(chapter).Error; err != nil {
					return fmt.Errorf("failed to update chapter order: %w", err)
				}
			}
		}

		return nil
	})
}

// GetNavbarOrder возвращает порядок элементов в навбаре по страницам
func (s *Service) GetNavbarOrder() (map[enums.Page][]models.Chapter, error) {
	chapters, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[enums.Page][]models.Chapter)
	for _, chapter := range chapters {
		result[chapter.Page] = append(result[chapter.Page], chapter)
	}

	// Сортируем каждую страницу по BarIdx
	for page := range result {
		// Используем пузырьковую сортировку для простоты
		chapters := result[page]
		for i := 0; i < len(chapters); i++ {
			for j := i + 1; j < len(chapters); j++ {
				if chapters[i].BarIdx > chapters[j].BarIdx {
					chapters[i], chapters[j] = chapters[j], chapters[i]
				}
			}
		}
		result[page] = chapters
	}

	return result, nil
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}
