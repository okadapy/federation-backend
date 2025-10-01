// crud-controller.go
package crud

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller[T any] struct {
	service *Service[T]
	logger  *log.Logger
}

// NewCrudController создает новый экземпляр CrudController
func NewCrudController[T any](db *gorm.DB, logger *log.Logger) *Controller[T] {
	return &Controller[T]{
		service: NewCrudService[T](db, logger),
		logger:  logger,
	}
}

// Create обрабатывает POST запросы для создания новой сущности
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

// Get обрабатывает GET запросы для получения сущности по ID
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

// GetAll обрабатывает GET запросы для получения всех сущностей
func (c *Controller[T]) GetAll(ctx *gin.Context) {
	entities, err := c.service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, entities)
}

// Update обрабатывает PUT запросы для обновления сущности по ID
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

// Delete обрабатывает DELETE запросы для удаления сущности по ID
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

// GetWithConditions обрабатывает GET запросы с условиями фильтрации
func (c *Controller[T]) GetWithConditions(ctx *gin.Context) {
	// Парсим параметры запроса
	includes := ctx.Query("includes")
	order := ctx.Query("order")
	limitStr := ctx.Query("limit")

	var limit int
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// Создаем карту условий из query параметров
	where := make(map[string]interface{})
	queryParams := ctx.Request.URL.Query()

	for key, values := range queryParams {
		// Пропускаем служебные параметры
		if key == "includes" || key == "order" || key == "limit" {
			continue
		}

		if len(values) > 0 {
			where[key] = values[0]
		}
	}

	var includeRelations []string
	if includes != "" {
		includeRelations = strings.Split(includes, ",")
	}

	entities, err := c.service.GetWithConditions(
		ctx.Request.Context(),
		where,
		includeRelations,
		order,
		limit,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, entities)
}
