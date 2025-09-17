package user

import (
	"federation-backend/app/api/shared"
	"federation-backend/app/db/models"

	"gorm.io/gorm"
)

type Controller struct {
	CRUD *shared.CrudController[models.User]
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{
		CRUD: shared.NewCrudController[models.User](db),
	}
}
