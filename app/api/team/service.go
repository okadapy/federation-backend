// team/service.go
package team

import (
	"errors"
	files "federation-backend/app/api/file"
	"federation-backend/app/db/models"
	"federation-backend/app/db/models/enums"
	"fmt"
	"mime/multipart"

	"gorm.io/gorm"
)

type CreateTeamDTO struct {
	TeamName string                `form:"teamName" binding:"required"`
	Sex      enums.Sex             `form:"sex" binding:"required"`
	TeamLogo *multipart.FileHeader `form:"teamLogo" binding:"required"`
}

type UpdateTeamDTO struct {
	TeamName *string               `form:"teamName"`
	Sex      *enums.Sex            `form:"sex"`
	TeamLogo *multipart.FileHeader `form:"teamLogo"`
}

type Service struct {
	db *gorm.DB
	fs *files.Service
}

func (s *Service) Create(dto interface{}) error {
	createDTO, ok := dto.(*CreateTeamDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	logo, err := s.fs.SaveFile(createDTO.TeamLogo)
	if err != nil {
		return err
	}

	team := models.Team{
		TeamName:   createDTO.TeamName,
		Sex:        createDTO.Sex,
		TeamLogoID: logo.Id,
	}

	if err := s.db.Create(&team).Error; err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

func (s *Service) Get(id uint) (models.Team, error) {
	var team models.Team
	err := s.db.Preload("TeamLogo").First(&team, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Team{}, errors.New("team not found")
		}
		return models.Team{}, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

func (s *Service) Update(id uint, dto interface{}) error {
	updateDTO, ok := dto.(*UpdateTeamDTO)
	if !ok {
		return errors.New("invalid DTO type")
	}

	var team models.Team
	if err := s.db.First(&team, id).Error; err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	if updateDTO.TeamName != nil {
		team.TeamName = *updateDTO.TeamName
	}
	if updateDTO.Sex != nil {
		team.Sex = *updateDTO.Sex
	}
	if updateDTO.TeamLogo != nil {

	}

	if err := s.db.Save(&team).Error; err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	return nil
}

func (s *Service) Delete(id uint) error {
	var team models.Team
	if err := s.db.First(&team, id).Error; err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	if err := s.db.Delete(&team).Error; err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

func (s *Service) GetAll() ([]models.Team, error) {
	var teams []models.Team
	if err := s.db.Preload("TeamLogo").Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	return teams, nil
}

func NewService(db *gorm.DB, fs *files.Service) *Service {
	return &Service{
		db: db,
		fs: fs,
	}
}
