package interfaces

import "github.com/gin-gonic/gin"

type Controller[Dto interface{}] interface {
	Create(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Get(ctx *gin.Context)
	GetAll(ctx *gin.Context)
	Update(ctx *gin.Context)
}
