package team

import (
	"federation-backend/app/api"
	"federation-backend/app/db/models"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	Service api.Service[models.Team]
}

func NewController(service api.Service[models.Team]) *Controller {
	return &Controller{
		Service: service,
	}
}

func (c *Controller) Create(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c *Controller) Delete(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c *Controller) Get(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c *Controller) GetAll(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (c *Controller) Update(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}
