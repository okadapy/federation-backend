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

func (s *Service) SaveFile(fileHeader *multipart.FileHeader) (string, error) {
	fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !isAllowedExtension(fileExt) {
		return "", errors.New("disallowed file extension")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	fileID := uuid.New().String()
	filename := fileID + fileExt
	path := filepath.Join(s.storagePath, filename)

	dst, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(path) // Clean up
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	metadata := models.File{
		Name: fileHeader.Filename,
		Size: fileHeader.Size,
		Path: path,
	}
	if err := s.db.Create(&metadata).Error; err != nil {
		os.Remove(path) // Clean up
		return "", fmt.Errorf("failed to save metadata: %w", err)
	}

	return filename, nil
}

func (s *Service) GetFile(filename string) (io.ReadCloser, error) {
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return nil, errors.New("invalid filename")
	}
	path := filepath.Join(s.storagePath, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (s *Service) DeleteFile(filename string) error {
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return errors.New("invalid filename")
	}
	path := filepath.Join(s.storagePath, filename)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func isAllowedExtension(ext string) bool {
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".pdf":  true,
	}
	return allowed[ext]
}
