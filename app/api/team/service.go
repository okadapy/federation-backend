package team

import (
	"errors"
	files "federation-backend/app/api/file"
	"federation-backend/app/db/models"
	"federation-backend/app/db/models/enums"
	"fmt"
	"mime/multipart"
	"path/filepath"

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

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Save the logo file
		logo, err := s.fs.SaveFile(createDTO.TeamLogo)
		if err != nil {
			return fmt.Errorf("failed to save team logo: %w", err)
		}

		team := models.Team{
			TeamName:   createDTO.TeamName,
			Sex:        createDTO.Sex,
			TeamLogoID: logo.Id, // Assuming ID field, not Id
		}

		if err := tx.Create(&team).Error; err != nil {
			// Clean up the saved file if team creation fails
			if deleteErr := s.fs.DeleteFile(filepath.Base(logo.Path)); deleteErr != nil {
				fmt.Printf("Warning: failed to clean up logo file after team creation failure: %v\n", deleteErr)
			}
			return fmt.Errorf("failed to create team: %w", err)
		}

		return nil
	})
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

	return s.db.Transaction(func(tx *gorm.DB) error {
		var team models.Team
		if err := tx.Preload("TeamLogo").First(&team, id).Error; err != nil {
			return fmt.Errorf("team not found: %w", err)
		}

		// Update basic fields
		if updateDTO.TeamName != nil {
			team.TeamName = *updateDTO.TeamName
		}
		if updateDTO.Sex != nil {
			team.Sex = *updateDTO.Sex
		}

		// Handle logo update if provided
		if updateDTO.TeamLogo != nil {
			// Save new logo first
			newLogo, err := s.fs.SaveFile(updateDTO.TeamLogo)
			if err != nil {
				return fmt.Errorf("failed to save new team logo: %w", err)
			}

			// Store the old logo ID for cleanup
			oldLogoID := team.TeamLogoID

			// Update team with new logo
			team.TeamLogoID = newLogo.Id

			// Save the team
			if err := tx.Save(&team).Error; err != nil {
				// Clean up the new logo if team update fails
				if deleteErr := s.fs.DeleteFile(filepath.Base(newLogo.Path)); deleteErr != nil {
					fmt.Printf("Warning: failed to clean up new logo file after update failure: %v\n", deleteErr)
				}
				return fmt.Errorf("failed to update team: %w", err)
			}

			// Delete old logo if it exists
			if oldLogoID != 0 {
				var oldLogo models.File
				if err := tx.First(&oldLogo, oldLogoID).Error; err == nil {
					// Extract filename from path for deletion
					filename := filepath.Base(oldLogo.Path)
					if deleteErr := s.fs.DeleteFile(filename); deleteErr != nil {
						fmt.Printf("Warning: failed to delete old logo file: %v\n", deleteErr)
					}
					// Delete the old file record
					if deleteErr := tx.Delete(&oldLogo).Error; deleteErr != nil {
						fmt.Printf("Warning: failed to delete old logo record: %v\n", deleteErr)
					}
				}
			}
		} else {
			// No logo update, just save the team
			if err := tx.Save(&team).Error; err != nil {
				return fmt.Errorf("failed to update team: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var team models.Team
		if err := tx.Preload("TeamLogo").First(&team, id).Error; err != nil {
			return fmt.Errorf("team not found: %w", err)
		}

		// Delete the team logo if it exists
		if team.TeamLogoID != 0 {
			var logo models.File
			if err := tx.First(&logo, team.TeamLogoID).Error; err == nil {
				// Extract filename from path for deletion
				filename := filepath.Base(logo.Path)
				if deleteErr := s.fs.DeleteFile(filename); deleteErr != nil {
					fmt.Printf("Warning: failed to delete team logo file: %v\n", deleteErr)
				}
				// Delete the file record
				if deleteErr := tx.Delete(&logo).Error; deleteErr != nil {
					fmt.Printf("Warning: failed to delete logo record: %v\n", deleteErr)
				}
			}
		}

		// Delete the team
		if err := tx.Delete(&team).Error; err != nil {
			return fmt.Errorf("failed to delete team: %w", err)
		}

		return nil
	})
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
