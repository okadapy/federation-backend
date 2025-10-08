// match.go
package models

import (
	"federation-backend/app/db/models/enums"
	"time"
)

type Match struct {
	Model
	League string    `json:"league" gorm:"size:100"`
	Date   time.Time `json:"date"`
	Sex    enums.Sex `json:"sex" gorm:"type:ENUM('female','male')"`
	Teams  []*Team   `json:"teams" gorm:"many2many:match_teams;"`
}
