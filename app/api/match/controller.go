package match

import (
	"federation-backend/app/db/models"
	"federation-backend/app/interfaces"
)

type Controller struct {
	srv *interfaces.Service[models.Match]
}
