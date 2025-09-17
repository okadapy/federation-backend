package models

type CallbackType string

const (
	TeamApplication CallbackType = "team_application"
	CallbackRequest              = "callback_request"
)

type CallBack struct {
	Model
	Name         string       `json:"name"`
	Phone        string       `json:"phone"`
	Email        *string      `json:"email"`
	TeamName     *string      `json:"team_name"`
	CallbackType CallbackType `json:"callback_type" gorm:"default:'callback_request'"`
}
