package repository

import (
	"context"
	"fmt"

	"personal-digital-inventory/internal/domain"

	"gorm.io/gorm"
)

type GormIgnorePatternRepository struct {
	db *gorm.DB
}

func NewGormIgnorePatternRepository(db *gorm.DB) *GormIgnorePatternRepository {
	return &GormIgnorePatternRepository{db: db}
}

func (r *GormIgnorePatternRepository) Create(ctx context.Context, pattern *domain.IgnorePattern) error {
	model := IgnorePatternFromDomain(pattern)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create ignore pattern: %w", result.Error)
	}
	pattern.ID = model.ID
	return nil
}

func (r *GormIgnorePatternRepository) FindByID(ctx context.Context, id uint) (*domain.IgnorePattern, error) {
	var model IgnorePatternModel
	result := r.db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find ignore pattern: %w", result.Error)
	}
	return model.ToDomain(), nil
}

func (r *GormIgnorePatternRepository) List(ctx context.Context) ([]domain.IgnorePattern, error) {
	var models []IgnorePatternModel
	result := r.db.WithContext(ctx).Order("id ASC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list ignore patterns: %w", result.Error)
	}

	patterns := make([]domain.IgnorePattern, len(models))
	for i, m := range models {
		patterns[i] = *m.ToDomain()
	}
	return patterns, nil
}

func (r *GormIgnorePatternRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&IgnorePatternModel{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete ignore pattern: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ignore pattern with id %d not found", id)
	}
	return nil
}
