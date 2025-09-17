package models

import "federation-backend/app/db/models/enums"

type Team struct {
	Model
	TeamName   string    `json:"team_name"`
	Sex        enums.Sex `json:"sex" gorm:"default:'male'"`
	TeamLogoID uint64    `json:"team_logo_id"`
	TeamLogo   FilePath  `gorm:"foreignkey:TeamLogoID"`
}
