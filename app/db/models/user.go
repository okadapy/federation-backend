package models

type User struct {
	Model
	Username string `json:"username" gorm:"uniqueIndex"`
	Password string `json:"password"`
}
