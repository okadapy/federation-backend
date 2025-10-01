// user.go
package models

type User struct {
	Model
	Username string `json:"username" gorm:"uniqueIndex;size:255"`
	Password string `json:"password" gorm:"size:255"`
}
