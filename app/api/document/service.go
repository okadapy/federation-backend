package document

import (
	"errors"
	"federation-backend/app/db/models"
	"federation-backend/app/db/models/enums"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"gorm.io/gorm"
)

type CreateDocumentDTO struct {
	Name    string                `form:"name" binding:"required"`
	Chapter enums.Doctype         `form:"chapter" binding:"required"`
	File    *multipart.FileHeader `form:"file" binding:"required"`
}

type UpdateDocumentDTO struct {
	Name    *string               `form:"name"`
	Chapter *enums.Doctype        `form:"chapter"`
	File    *multipart.FileHeader `form:"file"`
}

type Service struct {
	db          *gorm.DB
	fileService FileService
}

type FileService interface {
	SaveFile(fileHeader *multipart.FileHeader) (*models.File, error)
	DeleteFile(filename string) error
}

func (s *Service) Create(dto interface{}) error {
	createDTO, ok := dto.(*CreateDocumentDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Save the file first
		file, err := s.fileService.SaveFile(createDTO.File)
		if err != nil {
			return fmt.Errorf("failed to save document file: %w", err)
		}

		// Create the document with embedded File and additional fields
		document := models.Document{
			File:    *file, // Embed the File struct
			Name:    createDTO.Name,
			Chapter: createDTO.Chapter,
		}

		if err := tx.Create(&document).Error; err != nil {
			// Clean up the saved file if document creation fails
			if deleteErr := s.fileService.DeleteFile(filepath.Base(file.Path)); deleteErr != nil {
				fmt.Printf("Warning: failed to clean up document file after creation failure: %v\n", deleteErr)
			}
			return fmt.Errorf("failed to create document: %w", err)
		}

		return nil
	})
}

func (s *Service) Get(id uint) (models.Document, error) {
	var document models.Document
	err := s.db.First(&document, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Document{}, errors.New("document not found")
		}
		return models.Document{}, fmt.Errorf("failed to get document: %w", err)
	}
	return document, nil
}

func (s *Service) Update(id uint, dto interface{}) error {
	updateDTO, ok := dto.(*UpdateDocumentDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var document models.Document
		if err := tx.First(&document, id).Error; err != nil {
			return fmt.Errorf("document not found: %w", err)
		}

		// Update basic fields if provided
		if updateDTO.Name != nil {
			document.Name = *updateDTO.Name
		}
		if updateDTO.Chapter != nil {
			document.Chapter = *updateDTO.Chapter
		}

		// Handle file update if provided
		if updateDTO.File != nil {
			// Save new file first
			newFile, err := s.fileService.SaveFile(updateDTO.File)
			if err != nil {
				return fmt.Errorf("failed to save new document file: %w", err)
			}

			// Store the old file path for cleanup
			oldFilePath := document.File.Path

			// Update document with new file data
			document.File = *newFile

			// Save the document
			if err := tx.Save(&document).Error; err != nil {
				// Clean up the new file if document update fails
				if deleteErr := s.fileService.DeleteFile(filepath.Base(newFile.Path)); deleteErr != nil {
					fmt.Printf("Warning: failed to clean up new document file after update failure: %v\n", deleteErr)
				}
				return fmt.Errorf("failed to update document: %w", err)
			}

			// Delete old file
			if oldFilePath != "" {
				filename := filepath.Base(oldFilePath)
				if deleteErr := s.fileService.DeleteFile(filename); deleteErr != nil {
					fmt.Printf("Warning: failed to delete old document file: %v\n", deleteErr)
				}
			}
		} else {
			// No file update, just save the document
			if err := tx.Save(&document).Error; err != nil {
				return fmt.Errorf("failed to update document: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var document models.Document
		if err := tx.First(&document, id).Error; err != nil {
			return fmt.Errorf("document not found: %w", err)
		}

		// Delete the associated file
		if document.File.Path != "" {
			filename := filepath.Base(document.File.Path)
			if err := s.fileService.DeleteFile(filename); err != nil {
				fmt.Printf("Warning: failed to delete document file: %v\n", err)
			}
		}

		// Delete the document record
		if err := tx.Delete(&document).Error; err != nil {
			return fmt.Errorf("failed to delete document: %w", err)
		}

		return nil
	})
}

func (s *Service) GetAll() ([]models.Document, error) {
	var documents []models.Document
	err := s.db.Find(&documents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	return documents, nil
}

func (s *Service) GetByChapter(chapter enums.Doctype) ([]models.Document, error) {
	var documents []models.Document
	err := s.db.Where("chapter = ?", chapter).Find(&documents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get documents by chapter: %w", err)
	}
	return documents, nil
}

func NewService(db *gorm.DB, fileService FileService) *Service {
	return &Service{
		db:          db,
		fileService: fileService,
	}
}
