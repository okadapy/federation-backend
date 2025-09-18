package gallery_item

import (
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func (c *Controller) Create(ctx *gin.Context) {
	var dto CreateGalleryItemDTO
	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Create(&dto); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (c *Controller) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := c.service.Get(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item)
}

func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var dto UpdateGalleryItemDTO
	if err := ctx.ShouldBind(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Update(uint(id), &dto); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (c *Controller) GetAll(ctx *gin.Context) {
	// Implementation would depend on specific requirements
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func NewController(db *gorm.DB, service FileService) *Controller {
	return &Controller{
		service: NewService(db, service),
	}
}
