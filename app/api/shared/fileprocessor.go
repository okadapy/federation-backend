package shared

import (
	"log"
	"mime/multipart"
	"sync"

	files "federation-backend/app/api/file"
	"federation-backend/app/db/models"
)

type FileProcessor interface {
	SaveFile(fileHeader *multipart.FileHeader) (*models.File, error)
	DeleteFile(filename string) error
	SaveFilesParallel(files []*multipart.FileHeader) ([]*models.File, []error)
}

type ConcurrentFileProcessor struct {
	fileService *files.Service
	logger      *log.Logger
}

func NewConcurrentFileProcessor(fileService *files.Service, logger *log.Logger) *ConcurrentFileProcessor {
	return &ConcurrentFileProcessor{
		fileService: fileService,
		logger:      logger,
	}
}

func (p *ConcurrentFileProcessor) SaveFilesParallel(files []*multipart.FileHeader) ([]*models.File, []error) {
	var wg sync.WaitGroup
	results := make([]*models.File, len(files))
	errors := make([]error, len(files))

	for i, file := range files {
		p.logger.Println("[%d] Saving image...", i)
		wg.Add(1)
		go func(idx int, f *multipart.FileHeader) {
			defer wg.Done()
			file, err := p.fileService.SaveFile(f)
			results[idx] = file
			errors[idx] = err
			if err != nil {
				p.logger.Println("[%d] Image wasn't saved!")
			} else {
				p.logger.Println("[%d] Image was saved!")
			}
		}(i, file)
	}

	wg.Wait()
	return results, errors
}

// Реализуем остальные методы интерфейса
func (p *ConcurrentFileProcessor) SaveFile(fileHeader *multipart.FileHeader) (*models.File, error) {
	return p.fileService.SaveFile(fileHeader)
}

func (p *ConcurrentFileProcessor) DeleteFile(filename string) error {
	return p.fileService.DeleteFile(filename)
}
