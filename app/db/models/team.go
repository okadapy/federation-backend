// team.go
package models

import "federation-backend/app/db/models/enums"

type Team struct {
	Model
	TeamName   string    `json:"team_name" gorm:"size:255"`
	Sex        enums.Sex `json:"sex" gorm:"default:'male'"`
	TeamLogoID int       `json:"team_logo_id"`
	TeamLogo   File      `gorm:"foreignkey:TeamLogoID" json:"logo"`
}
