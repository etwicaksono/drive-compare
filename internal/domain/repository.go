package domain

import "context"

type ComparisonRepository interface {
	Create(ctx context.Context, comparison *Comparison) error
	FindByID(ctx context.Context, id uint) (*Comparison, error)
	List(ctx context.Context) ([]Comparison, error)
	Delete(ctx context.Context, id uint) error
}

type IgnorePatternRepository interface {
	Create(ctx context.Context, pattern *IgnorePattern) error
	FindByID(ctx context.Context, id uint) (*IgnorePattern, error)
	List(ctx context.Context) ([]IgnorePattern, error)
	Delete(ctx context.Context, id uint) error
}
