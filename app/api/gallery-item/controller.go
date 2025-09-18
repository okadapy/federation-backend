package gallery_item

import (
	"federation-backend/app/db/models"
	"federation-backend/app/interfaces"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	interfaces.Service[models.GalleryItem]
}

func (c Controller) Create(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c Controller) Delete(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c Controller) Get(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c Controller) GetAll(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c Controller) Update(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}
