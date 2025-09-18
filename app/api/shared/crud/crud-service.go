package crud

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Service[T any] struct {
	db *gorm.DB
}

func NewCrudService[T any](db *gorm.DB) *Service[T] {
	return &Service[T]{db: db}
}

func (c *Service[T]) Create(ctx context.Context, dto *T) error {
	if dto == nil {
		return errors.New("dto cannot be nil")
	}

	result := c.db.WithContext(ctx).Create(dto)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *Service[T]) CreateInBatches(ctx context.Context, dtos []*T, batchSize int) error {
	if len(dtos) == 0 {
		return errors.New("no records to create")
	}

	result := c.db.WithContext(ctx).CreateInBatches(dtos, batchSize)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *Service[T]) Get(ctx context.Context, id uint) (*T, error) {
	var model T
	result := c.db.WithContext(ctx).First(&model, id)
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
	result := c.db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("record not found")
		}
		return result.Error
	}

	result = c.db.WithContext(ctx).Model(&model).Updates(dto)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *Service[T]) Delete(ctx context.Context, id uint) error {
	var model T
	result := c.db.WithContext(ctx).Delete(&model, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

func (c *Service[T]) GetAll(ctx context.Context, whereClause string, args []interface{}) ([]T, error) {
	var models []T
	query := c.db.WithContext(ctx).Model(new(T))

	if whereClause != "" {
		query = query.Where(whereClause, args...)
	}

	result := query.Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	return models, nil
}

func (c *Service[T]) GetWithConditions(
	ctx context.Context,
	where map[string]interface{},
	includes []string,
	order string,
	limit int,
) ([]T, error) {
	var models []T
	query := c.db.WithContext(ctx).Model(new(T))

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
