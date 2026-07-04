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

func (r *GormComparisonRepository) SaveFiles(ctx context.Context, files []domain.ComparisonFile) error {
	if len(files) == 0 {
		return nil
	}

	models := make([]ComparisonFileModel, len(files))
	for i, f := range files {
		models[i] = ComparisonFileModel{
			ComparisonID: f.ComparisonID,
			Filename:     f.Filename,
			Extension:    f.Extension,
			RelativePath: f.RelativePath,
			Size:         f.Size,
			Status:       string(f.Status),
		}
	}

	result := r.db.WithContext(ctx).Create(&models)
	if result.Error != nil {
		return fmt.Errorf("failed to save comparison files: %w", result.Error)
	}
	return nil
}

func (r *GormComparisonRepository) ListFilesByComparisonID(ctx context.Context, comparisonID uint) ([]domain.ComparisonFile, error) {
	var models []ComparisonFileModel
	result := r.db.WithContext(ctx).Where("comparison_id = ?", comparisonID).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list comparison files: %w", result.Error)
	}

	files := make([]domain.ComparisonFile, len(models))
	for i, m := range models {
		files[i] = domain.ComparisonFile{
			ID:           m.ID,
			ComparisonID: m.ComparisonID,
			Filename:     m.Filename,
			Extension:    m.Extension,
			RelativePath: m.RelativePath,
			Size:         m.Size,
			Status:       domain.FileStatus(m.Status),
		}
	}
	return files, nil
}
