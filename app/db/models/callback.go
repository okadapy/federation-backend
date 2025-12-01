// callback.go
package models

type CallbackType string

const (
	TeamApplication CallbackType = "team_application"
	CallbackRequest              = "callback_request"
)

type CallBack struct {
	Model
	Name         string       `json:"name" gorm:"size:100"`
	Phone        string       `json:"phone" gorm:"size:20"`
	Email        *string      `json:"email" gorm:"size:255"`
	TeamName     *string      `json:"team_name" gorm:"size:255"`
	CallbackType CallbackType `json:"callback_type"`
}
