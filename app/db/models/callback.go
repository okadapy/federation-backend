package models

type CallBack struct {
	Model
	Name  string  `json:"name"`
	Phone string  `json:"phone" gorm:"uniqueIndex"`
	Email *string `json:"email" gorm:"uniqueIndex"`
}
