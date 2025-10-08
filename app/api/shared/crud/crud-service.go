// crud-service.go
package crud

import (
	"context"
	"errors"
	"log"

	"gorm.io/gorm"
)

type Service[T any] struct {
	Db     *gorm.DB
	logger *log.Logger
}

func NewCrudService[T any](db *gorm.DB, logger *log.Logger) *Service[T] {
	return &Service[T]{Db: db, logger: logger}
}

func (c *Service[T]) Create(ctx context.Context, dto *T) error {
	if dto == nil {
		return errors.New("dto cannot be nil")
	}

	result := c.Db.WithContext(ctx).Create(dto)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *Service[T]) CreateInBatches(ctx context.Context, dtos []*T, batchSize int) error {
	if len(dtos) == 0 {
		return errors.New("no records to create")
	}

	result := c.Db.WithContext(ctx).CreateInBatches(dtos, batchSize)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *Service[T]) Get(ctx context.Context, id uint) (*T, error) {
	var model T
	result := c.Db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &model, nil
}

func (c *Service[T]) Update(ctx context.Context, id uint, dto *T) error {
	var model T
	result := c.Db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("record not found")
		}
		return result.Error
	}

	result = c.Db.WithContext(ctx).Model(&model).Updates(dto)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *Service[T]) Delete(ctx context.Context, id uint) error {
	var model T
	result := c.Db.WithContext(ctx).Delete(&model, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

func (c *Service[T]) GetWithConditions(
	ctx context.Context,
	where map[string]interface{},
	includes []string,
	order string,
	limit int,
) ([]T, error) {
	var models []T
	query := c.Db.WithContext(ctx).Model(new(T))

	if len(where) > 0 {
		query = query.Where(where)
	}

	for _, include := range includes {
		query = query.Preload(include)
	}

	if order != "" {
		query = query.Order(order)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	return models, nil
}

// GetAll возвращает все записи с возможностью предзагрузки связей
func (c *Service[T]) GetAll() ([]T, error) {
	var models []T
	defer c.logger.Printf("Exiting GetAll()")
	c.logger.Printf("GetAll() invoked for %+v\n", *c)
	result := c.Db.Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("record not found")
	}
	return models, nil
}

// UpdateWithAssociations обновляет запись со связями
func (c *Service[T]) UpdateWithAssociations(ctx context.Context, id uint, dto *T) error {
	var model T
	result := c.Db.First(&model, id)
	if result.Error != nil {
		return result.Error
	}

	result = c.Db.WithContext(ctx).Model(&model).Save(dto)
	return result.Error
}
