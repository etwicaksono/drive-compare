package repository

import (
	"time"

	"personal-digital-inventory/internal/domain"
)

type ComparisonModel struct {
	ID             uint       `gorm:"primaryKey"`
	LocalPath      string     `gorm:"column:local_path;not null"`
	DrivePath      string     `gorm:"column:drive_path;not null"`
	StartedAt      time.Time  `gorm:"column:started_at;not null"`
	FinishedAt     *time.Time `gorm:"column:finished_at"`
	TotalLocalOnly int        `gorm:"column:total_local_only;not null;default:0"`
	TotalDriveOnly int        `gorm:"column:total_drive_only;not null;default:0"`
	TotalBoth      int        `gorm:"column:total_both;not null;default:0"`
}

func (ComparisonModel) TableName() string {
	return "comparisons"
}

func (m *ComparisonModel) ToDomain() *domain.Comparison {
	return &domain.Comparison{
		ID:             m.ID,
		LocalPath:      m.LocalPath,
		DrivePath:      m.DrivePath,
		StartedAt:      m.StartedAt,
		FinishedAt:     m.FinishedAt,
		TotalLocalOnly: m.TotalLocalOnly,
		TotalDriveOnly: m.TotalDriveOnly,
		TotalBoth:      m.TotalBoth,
	}
}

func ComparisonFromDomain(c *domain.Comparison) *ComparisonModel {
	return &ComparisonModel{
		ID:             c.ID,
		LocalPath:      c.LocalPath,
		DrivePath:      c.DrivePath,
		StartedAt:      c.StartedAt,
		FinishedAt:     c.FinishedAt,
		TotalLocalOnly: c.TotalLocalOnly,
		TotalDriveOnly: c.TotalDriveOnly,
		TotalBoth:      c.TotalBoth,
	}
}

type ComparisonFileModel struct {
	ID           uint   `gorm:"primaryKey"`
	ComparisonID uint   `gorm:"column:comparison_id;not null"`
	Filename     string `gorm:"column:filename;not null"`
	Extension    string `gorm:"column:extension;not null;default:''"`
	RelativePath string `gorm:"column:relative_path;not null"`
	Size         int64  `gorm:"column:size;not null;default:0"`
	Status       string `gorm:"column:status;not null"`
}

func (ComparisonFileModel) TableName() string {
	return "comparison_files"
}

type IgnorePatternModel struct {
	ID      uint   `gorm:"primaryKey"`
	Pattern string `gorm:"column:pattern;not null;uniqueIndex"`
}

func (IgnorePatternModel) TableName() string {
	return "ignore_patterns"
}

func (m *IgnorePatternModel) ToDomain() *domain.IgnorePattern {
	return &domain.IgnorePattern{
		ID:      m.ID,
		Pattern: m.Pattern,
	}
}

func IgnorePatternFromDomain(p *domain.IgnorePattern) *IgnorePatternModel {
	return &IgnorePatternModel{
		ID:      p.ID,
		Pattern: p.Pattern,
	}
}
