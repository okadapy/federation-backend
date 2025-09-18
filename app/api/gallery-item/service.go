package gallery_item

import (
	files "federation-backend/app/api/file"
	"federation-backend/app/db/models"

	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	fileService *files.Service
}

func (s Service) Create(dto interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Get(id uint) (models.GalleryItem, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) Update(id uint, dto interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Delete(id uint) error {
	//TODO implement me
	panic("implement me")
}

func NewService(db *gorm.DB, fileService *files.Service) *Service {
	return &Service{
		db:          db,
		fileService: fileService,
	}
}
