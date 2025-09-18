package crud

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller[T any] struct {
	service *Service[T]
}

// NewCrudController creates a new instance of CrudController
func NewCrudController[T any](db *gorm.DB) *Controller[T] {
	return &Controller[T]{
		service: NewCrudService[T](db),
	}
}

// Create handles POST requests to create a new entity
func (c *Controller[T]) Create(ctx *gin.Context) {
	var dto T
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Create(ctx.Request.Context(), &dto); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto)
}

// Get handles GET requests to retrieve an entity by ID
func (c *Controller[T]) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	entity, err := c.service.Get(ctx.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, entity)
}

// GetAll handles GET requests to retrieve all entities
func (c *Controller[T]) GetAll(ctx *gin.Context) {
	// Parse query parameters for filtering
	queryParams := ctx.Request.URL.Query()
	whereClause := ""
	var args []interface{}

	// You can extend this to handle specific query parameters
	// For example: ?age=30 could become "age = ?" with arg 30
	for key, values := range queryParams {
		if len(values) > 0 {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += key + " = ?"
			args = append(args, values[0])
		}
	}

	entities, err := c.service.GetAll(ctx.Request.Context(), whereClause, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, entities)
}

// Update handles PUT requests to update an entity by ID
func (c *Controller[T]) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var dto T
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Update(ctx.Request.Context(), uint(id), &dto); err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto)
}

// Delete handles DELETE requests to remove an entity by ID
func (c *Controller[T]) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.service.Delete(ctx.Request.Context(), uint(id)); err != nil {
		if err.Error() == "record not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Entity deleted successfully"})
}
