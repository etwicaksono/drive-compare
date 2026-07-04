package repository

import (
	"context"
	"fmt"

	"personal-digital-inventory/internal/domain"

	"gorm.io/gorm"
)

type GormComparisonRepository struct {
	db *gorm.DB
}

func NewGormComparisonRepository(db *gorm.DB) *GormComparisonRepository {
	return &GormComparisonRepository{db: db}
}

func (r *GormComparisonRepository) Create(ctx context.Context, comparison *domain.Comparison) error {
	model := ComparisonFromDomain(comparison)
	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create comparison: %w", result.Error)
	}
	comparison.ID = model.ID
	return nil
}

func (r *GormComparisonRepository) FindByID(ctx context.Context, id uint) (*domain.Comparison, error) {
	var model ComparisonModel
	result := r.db.WithContext(ctx).First(&model, id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find comparison: %w", result.Error)
	}
	return model.ToDomain(), nil
}

func (r *GormComparisonRepository) List(ctx context.Context) ([]domain.Comparison, error) {
	var models []ComparisonModel
	result := r.db.WithContext(ctx).Order("id DESC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list comparisons: %w", result.Error)
	}

	comparisons := make([]domain.Comparison, len(models))
	for i, m := range models {
		comparisons[i] = *m.ToDomain()
	}
	return comparisons, nil
}

func (r *GormComparisonRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&ComparisonModel{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete comparison: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("comparison with id %d not found", id)
	}
	return nil
}
