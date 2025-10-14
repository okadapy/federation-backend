package files

import (
	"errors"
	"federation-backend/app/db/models"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	storagePath string
}

func NewService(db *gorm.DB, storagePath string) (*Service, error) {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &Service{db: db, storagePath: storagePath}, nil
}

func (s *Service) SaveFile(fileHeader *multipart.FileHeader) (*models.File, error) {
	fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !isAllowedExtension(fileExt) {
		return nil, errors.New("disallowed file extension")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	fileID := uuid.New().String()
	filename := fileID + fileExt
	path := filepath.Join(s.storagePath, filename)

	dst, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(path) // Clean up
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	metadata := models.File{
		Name: fileHeader.Filename,
		Size: fileHeader.Size,
		Path: filename, // Store only filename, not full path
	}
	if err := s.db.Create(&metadata).Error; err != nil {
		os.Remove(path) // Clean up
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return &metadata, nil
}

func (s *Service) DeleteFile(filename string) error {
	// Security check - prevent path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return errors.New("invalid filename")
	}

	path := filepath.Join(s.storagePath, filename)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, which might be ok in some cases
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Helper function to serve files
func (s *Service) GetFilePath(filename string) string {
	return filepath.Join(s.storagePath, filename)
}

func isAllowedExtension(ext string) bool {
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".txt":  true,
		".xlsx": true,
		".xls":  true,
	}
	return allowed[ext]
}
