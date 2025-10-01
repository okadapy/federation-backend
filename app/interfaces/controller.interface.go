package interfaces

import "github.com/gin-gonic/gin"

type Controller interface {
	Create(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Get(ctx *gin.Context)
	GetAll(ctx *gin.Context)
	Update(ctx *gin.Context)
}

func RegisterRoutes(c Controller, router *gin.RouterGroup) {
	router.GET("/", c.GetAll)
	router.GET("/:id", c.Get)
	router.POST("/", c.Create)
	router.PUT("/:id", c.Update)
	router.DELETE("/:id", c.Delete)
}
