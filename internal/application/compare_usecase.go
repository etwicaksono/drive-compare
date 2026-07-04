package application

import (
	"context"
	"fmt"
	"time"

	"personal-digital-inventory/internal/domain"
)

type LocalScannerFunc func(rootDir string) ([]domain.FileItem, error)
type DriveScannerFunc func(ctx context.Context, folderID string) ([]domain.FileItem, error)

type CompareUseCase struct {
	comparisonRepo domain.ComparisonRepository
	localScan      LocalScannerFunc
	driveScan      DriveScannerFunc
}

func NewCompareUseCase(
	repo domain.ComparisonRepository,
	localScan LocalScannerFunc,
	driveScan DriveScannerFunc,
) *CompareUseCase {
	return &CompareUseCase{
		comparisonRepo: repo,
		localScan:      localScan,
		driveScan:      driveScan,
	}
}

type CompareInput struct {
	LocalPath string
	DrivePath string
}

type CompareOutput struct {
	ComparisonID   uint
	LocalOnly      []domain.FileItem
	DriveOnly      []domain.FileItem
	Both           []domain.FileItem
	TotalLocalOnly int
	TotalDriveOnly int
	TotalBoth      int
}

func (uc *CompareUseCase) Execute(ctx context.Context, input CompareInput) (*CompareOutput, error) {
	if input.LocalPath == "" {
		return nil, fmt.Errorf("local path is required")
	}
	if input.DrivePath == "" {
		return nil, fmt.Errorf("drive path is required")
	}

	startedAt := time.Now()

	localFiles, err := uc.localScan(input.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("local scan failed: %w", err)
	}

	driveFiles, err := uc.driveScan(ctx, input.DrivePath)
	if err != nil {
		return nil, fmt.Errorf("drive scan failed: %w", err)
	}

	result := domain.CompareFiles(localFiles, driveFiles)

	finishedAt := time.Now()
	comparison := &domain.Comparison{
		LocalPath:      input.LocalPath,
		DrivePath:      input.DrivePath,
		StartedAt:      startedAt,
		FinishedAt:     &finishedAt,
		TotalLocalOnly: len(result.LocalOnly),
		TotalDriveOnly: len(result.DriveOnly),
		TotalBoth:      len(result.Both),
	}

	if err := uc.comparisonRepo.Create(ctx, comparison); err != nil {
		return nil, fmt.Errorf("failed to save comparison: %w", err)
	}

	var compFiles []domain.ComparisonFile
	compFiles = append(compFiles, toComparisonFiles(comparison.ID, result.LocalOnly, domain.FileStatusLocalOnly)...)
	compFiles = append(compFiles, toComparisonFiles(comparison.ID, result.DriveOnly, domain.FileStatusDriveOnly)...)
	compFiles = append(compFiles, toComparisonFiles(comparison.ID, result.Both, domain.FileStatusBoth)...)

	if err := uc.comparisonRepo.SaveFiles(ctx, compFiles); err != nil {
		return nil, fmt.Errorf("failed to save comparison files: %w", err)
	}

	return &CompareOutput{
		ComparisonID:   comparison.ID,
		LocalOnly:      result.LocalOnly,
		DriveOnly:      result.DriveOnly,
		Both:           result.Both,
		TotalLocalOnly: len(result.LocalOnly),
		TotalDriveOnly: len(result.DriveOnly),
		TotalBoth:      len(result.Both),
	}, nil
}

func toComparisonFiles(comparisonID uint, items []domain.FileItem, status domain.FileStatus) []domain.ComparisonFile {
	files := make([]domain.ComparisonFile, len(items))
	for i, item := range items {
		files[i] = domain.ComparisonFile{
			ComparisonID: comparisonID,
			Filename:     item.Name,
			Extension:    item.Extension,
			RelativePath: item.RelativePath,
			Size:         item.Size,
			Status:       status,
		}
	}
	return files
}
