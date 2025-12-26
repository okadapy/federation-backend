package gallery_item

import (
	"errors"
	"federation-backend/app/api/shared"
	"federation-backend/app/db/models"
	"fmt"
	"log"
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
	ChapterID     *uint                   `form:"chapter_id"`
	Name          *string                 `form:"name"`
	Date          *string                 `form:"date"`
	NewImages     []*multipart.FileHeader `form:"new_images"`
	OldImages     []int                   `form:"old_images"`
	DeletedImages []int                   `form:"deleted_images"`
	Preview       *multipart.FileHeader   `form:"preview"`
}

type Service struct {
	db          *gorm.DB
	fileService shared.FileProcessor
	logger      *log.Logger
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

func (s *Service) Create(createDTO *CreateGalleryItemDTO) error {
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
			PreviewID: preview.Id,
			Preview:   *preview,
		}

		if err := tx.Create(&galleryItem).Error; err != nil {
			return fmt.Errorf("failed to create gallery item: %w", err)
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
func (s *Service) Update(id uint, updateDTO *UpdateGalleryItemDTO) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Загружаем сущность
		item, err := s.loadGalleryItemWithAssociations(tx, id)
		if err != nil {
			return err
		}

		// Обновляем основные поля
		if err := s.updateBasicFields(tx, item, updateDTO); err != nil {
			return err
		}

		// Обновляем изображения
		if err := s.updateImages(tx, item, updateDTO); err != nil {
			return err
		}

		// Сохраняем изменения
		if err := tx.Save(item).Error; err != nil {
			return fmt.Errorf("failed to update gallery item: %w", err)
		}

		return nil
	})
}

// loadGalleryItemWithAssociations загружает галерею со всеми ассоциациями
func (s *Service) loadGalleryItemWithAssociations(tx *gorm.DB, id uint) (*models.GalleryItem, error) {
	var item models.GalleryItem
	if err := tx.Preload("Images").Preload("Preview").Preload("Chapter").
		First(&item, id).Error; err != nil {
		return nil, fmt.Errorf("gallery item not found: %w", err)
	}
	return &item, nil
}

// updateBasicFields обновляет основные поля галереи
func (s *Service) updateBasicFields(tx *gorm.DB, item *models.GalleryItem, dto *UpdateGalleryItemDTO) error {
	// Обновляем превью если предоставлено
	if dto.Preview != nil {
		if err := s.updatePreview(tx, item, dto.Preview); err != nil {
			return err
		}
	}

	// Обновляем основные поля
	if dto.ChapterID != nil {
		item.ChapterID = *dto.ChapterID

		// Загружаем новый Chapter для обновления ассоциации
		var chapter models.Chapter
		if err := tx.First(&chapter, *dto.ChapterID).Error; err != nil {
			return fmt.Errorf("failed to find chapter: %w", err)
		}

		// Обновляем ассоциацию
		if err := tx.Model(item).Association("Chapter").Replace(&chapter); err != nil {
			return fmt.Errorf("failed to update chapter association: %w", err)
		}
	}

	if dto.Name != nil {
		item.Name = *dto.Name
	}

	if dto.Date != nil {
		if err := s.updateDate(item, *dto.Date); err != nil {
			return err
		}
	}

	return nil
}

// updatePreview обновляет превью галереи
func (s *Service) updatePreview(tx *gorm.DB, item *models.GalleryItem, preview *multipart.FileHeader) error {
	if preview == nil {
		return nil
	}

	file, err := s.fileService.SaveFile(preview)
	if err != nil {
		return fmt.Errorf("failed to save preview: %w", err)
	}

	if err := tx.Model(item).Association("Preview").Replace(file); err != nil {
		return fmt.Errorf("failed to replace preview: %w", err)
	}

	return nil
}

// updateDate обновляет дату галереи
func (s *Service) updateDate(item *models.GalleryItem, dateStr string) error {
	var date time.Time
	if err := s.parseDate(dateStr, &date); err != nil {
		return err
	}
	item.Date = date
	return nil
}

// updateImages обрабатывает обновление изображений
func (s *Service) updateImages(tx *gorm.DB, item *models.GalleryItem, dto *UpdateGalleryItemDTO) error {
	// Удаляем помеченные изображения
	if err := s.deleteMarkedImages(tx, item, dto.DeletedImages); err != nil {
		return err
	}

	// Синхронизируем старые изображения
	if err := s.syncOldImages(tx, item, dto.OldImages, dto.DeletedImages); err != nil {
		return err
	}

	// Добавляем новые изображения
	if err := s.addNewImages(tx, item, dto.NewImages); err != nil {
		return err
	}

	return nil
}

// deleteMarkedImages удаляет изображения, помеченные для удаления
func (s *Service) deleteMarkedImages(tx *gorm.DB, item *models.GalleryItem, deletedImages []int) error {
	if len(deletedImages) == 0 {
		return nil
	}

	// Находим изображения для удаления
	var imagesToDelete []models.File
	if err := tx.Where("id IN ?", deletedImages).Find(&imagesToDelete).Error; err != nil {
		return fmt.Errorf("failed to find images to delete: %w", err)
	}

	// Удаляем каждое изображение
	for _, image := range imagesToDelete {
		if err := s.deleteSingleImage(tx, item, &image); err != nil {
			// Логируем ошибку, но продолжаем удаление остальных
			fmt.Printf("Warning: %v\n", err)
		}
	}

	return nil
}

// deleteSingleImage удаляет одно изображение
func (s *Service) deleteSingleImage(tx *gorm.DB, item *models.GalleryItem, image *models.File) error {
	// Удаляем из файловой системы
	filename := filepath.Base(image.Path)
	if err := s.fileService.DeleteFile(filename); err != nil {
		return fmt.Errorf("failed to delete image file %s: %w", filename, err)
	}

	// Удаляем из ассоциации
	if err := tx.Model(item).Association("Images").Delete(image); err != nil {
		return fmt.Errorf("failed to remove image %d from association: %w", image.Id, err)
	}

	// Удаляем запись из базы
	if err := tx.Delete(image).Error; err != nil {
		return fmt.Errorf("failed to delete file record %d: %w", image.Id, err)
	}

	return nil
}

// syncOldImages синхронизирует старые изображения
func (s *Service) syncOldImages(tx *gorm.DB, item *models.GalleryItem, oldImages []int, deletedImages []int) error {
	if oldImages == nil {
		return nil // Не обновляем старые изображения, если не указаны
	}

	// Получаем текущие изображения
	var currentImages []models.File
	if err := tx.Model(item).Association("Images").Find(&currentImages); err != nil {
		return fmt.Errorf("failed to get current images: %w", err)
	}

	// Идентифицируем изображения для удаления
	for _, currentImage := range currentImages {
		// Пропускаем если уже удалено
		if sliceContains(deletedImages, int(currentImage.Id)) {
			continue
		}

		// Проверяем нужно ли оставить изображение
		if !shouldKeepImage(&currentImage, oldImages) {
			if err := s.removeImageFromAssociation(tx, item, &currentImage); err != nil {
				return err
			}
		}
	}

	return nil
}

// shouldKeepImage определяет нужно ли оставлять изображение
func shouldKeepImage(image *models.File, oldImages []int) bool {
	for _, oldImageID := range oldImages {
		if int(image.Id) == oldImageID {
			return true
		}
	}
	return false
}

// removeImageFromAssociation удаляет изображение из ассоциации (но не из базы)
func (s *Service) removeImageFromAssociation(tx *gorm.DB, item *models.GalleryItem, image *models.File) error {
	if err := tx.Model(item).Association("Images").Delete(image); err != nil {
		return fmt.Errorf("failed to remove image %d from association: %w", image.Id, err)
	}
	return nil
}

// Обновляем метод addNewImages для параллельного сохранения
func (s *Service) addNewImages(tx *gorm.DB, item *models.GalleryItem, newImages []*multipart.FileHeader) error {
	if len(newImages) == 0 {
		return nil
	}

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
			if err := tx.Model(item).Association("Images").Append(file); err != nil {
				return fmt.Errorf("failed to associate image: %w", err)
			}
		}
	}

	return nil
}

// sliceContains проверяет наличие элемента в срезе
func sliceContains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
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

func NewService(db *gorm.DB, fileProcessor shared.FileProcessor, logger *log.Logger) *Service {
	return &Service{
		db:          db,
		fileService: fileProcessor,
		logger:      logger,
	}
}
