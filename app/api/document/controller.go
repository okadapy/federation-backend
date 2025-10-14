package document

import (
	"federation-backend/app/db/models/enums"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	service *Service
}

func (c *Controller) Create(ctx *gin.Context) {
	var dto CreateDocumentDTO
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

	document, err := c.service.Get(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, document)
}

func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var dto UpdateDocumentDTO
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
	documents, err := c.service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, documents)
}

func (c *Controller) GetByChapter(ctx *gin.Context) {
	chapterStr := ctx.Param("chapter")
	var chapter enums.Doctype

	// Convert string to Doctype enum
	switch chapterStr {
	case "rules":
		chapter = enums.Rules
	case "regulations":
		chapter = enums.Regulations
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter type"})
		return
	}

	documents, err := c.service.GetByChapter(chapter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, documents)
}

func NewController(db *gorm.DB, fileService FileService) *Controller {
	return &Controller{
		service: NewService(db, fileService),
	}
}
