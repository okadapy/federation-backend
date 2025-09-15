package models

import "federation-backend/app/db/models/enums"

type Team struct {
	Model

	Name            string    `json:"name" gorm:"uniqueIndex"`
	Sex             enums.Sex `json:"sex" gorm:"type:enum('male', 'female')"`
	LogoPathID      uint
	LogoPath        FilePath `json:"logo_path" gorm:"uniqueIndex, foreignKey:LogoPathID"`
	TeamSubmitterID uint
	TeamSubmitter   CallBack `json:"team_submitter" gorm:"foreignKey:TeamSubmitterID"`
}
