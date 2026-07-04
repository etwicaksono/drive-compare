package application

import (
	"context"
	"fmt"
	"testing"

	"personal-digital-inventory/internal/domain"
)

type mockComparisonRepo struct {
	comparison *domain.Comparison
	files      []domain.ComparisonFile
	createErr  error
	saveErr    error
}

func (m *mockComparisonRepo) Create(ctx context.Context, c *domain.Comparison) error {
	if m.createErr != nil {
		return m.createErr
	}
	c.ID = 1
	m.comparison = c
	return nil
}

func (m *mockComparisonRepo) FindByID(ctx context.Context, id uint) (*domain.Comparison, error) {
	return m.comparison, nil
}

func (m *mockComparisonRepo) List(ctx context.Context) ([]domain.Comparison, error) {
	return nil, nil
}

func (m *mockComparisonRepo) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockComparisonRepo) SaveFiles(ctx context.Context, files []domain.ComparisonFile) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.files = files
	return nil
}

func (m *mockComparisonRepo) ListFilesByComparisonID(ctx context.Context, comparisonID uint) ([]domain.ComparisonFile, error) {
	return m.files, nil
}

func TestCompareUseCase_Success(t *testing.T) {
	repo := &mockComparisonRepo{}

	localScan := func(ctx context.Context, rootDir string) ([]domain.FileItem, error) {
		return []domain.FileItem{
			{RelativePath: "", Name: "shared", Extension: ".txt", Size: 100},
			{RelativePath: "", Name: "local_only", Extension: ".doc", Size: 200},
		}, nil
	}

	driveScan := func(ctx context.Context, folderID string) ([]domain.FileItem, error) {
		return []domain.FileItem{
			{RelativePath: "", Name: "shared", Extension: ".txt", Size: 100},
			{RelativePath: "", Name: "drive_only", Extension: ".xls", Size: 300},
		}, nil
	}

	uc := NewCompareUseCase(repo, localScan, driveScan)
	output, err := uc.Execute(context.Background(), CompareInput{
		LocalPath: "D:\\Photos",
		DrivePath: "drive-folder-id",
	})

	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if output.TotalLocalOnly != 1 {
		t.Errorf("TotalLocalOnly = %d, want 1", output.TotalLocalOnly)
	}
	if output.TotalDriveOnly != 1 {
		t.Errorf("TotalDriveOnly = %d, want 1", output.TotalDriveOnly)
	}
	if output.TotalBoth != 1 {
		t.Errorf("TotalBoth = %d, want 1", output.TotalBoth)
	}
	if output.ComparisonID != 1 {
		t.Errorf("ComparisonID = %d, want 1", output.ComparisonID)
	}

	if repo.comparison == nil {
		t.Fatal("comparison not saved to repo")
	}
	if repo.comparison.LocalPath != "D:\\Photos" {
		t.Errorf("LocalPath = %q, want %q", repo.comparison.LocalPath, "D:\\Photos")
	}
	if len(repo.files) != 3 {
		t.Errorf("saved files = %d, want 3", len(repo.files))
	}
}

func TestCompareUseCase_EmptyLocalPath(t *testing.T) {
	uc := NewCompareUseCase(&mockComparisonRepo{}, nil, nil)
	_, err := uc.Execute(context.Background(), CompareInput{
		LocalPath: "",
		DrivePath: "drive-id",
	})

	if err == nil {
		t.Fatal("Execute() should return error for empty local path")
	}
}

func TestCompareUseCase_EmptyDrivePath(t *testing.T) {
	uc := NewCompareUseCase(&mockComparisonRepo{}, nil, nil)
	_, err := uc.Execute(context.Background(), CompareInput{
		LocalPath: "D:\\Photos",
		DrivePath: "",
	})

	if err == nil {
		t.Fatal("Execute() should return error for empty drive path")
	}
}

func TestCompareUseCase_LocalScanError(t *testing.T) {
	localScan := func(ctx context.Context, rootDir string) ([]domain.FileItem, error) {
		return nil, fmt.Errorf("directory not found")
	}
	driveScan := func(ctx context.Context, folderID string) ([]domain.FileItem, error) {
		return nil, nil
	}

	uc := NewCompareUseCase(&mockComparisonRepo{}, localScan, driveScan)
	_, err := uc.Execute(context.Background(), CompareInput{
		LocalPath: "/bad/path",
		DrivePath: "drive-id",
	})

	if err == nil {
		t.Fatal("Execute() should return error when local scan fails")
	}
}

func TestCompareUseCase_DriveScanError(t *testing.T) {
	localScan := func(ctx context.Context, rootDir string) ([]domain.FileItem, error) {
		return nil, nil
	}
	driveScan := func(ctx context.Context, folderID string) ([]domain.FileItem, error) {
		return nil, fmt.Errorf("API error")
	}

	uc := NewCompareUseCase(&mockComparisonRepo{}, localScan, driveScan)
	_, err := uc.Execute(context.Background(), CompareInput{
		LocalPath: "D:\\Photos",
		DrivePath: "drive-id",
	})

	if err == nil {
		t.Fatal("Execute() should return error when drive scan fails")
	}
}

func TestCompareUseCase_RepoCreateError(t *testing.T) {
	repo := &mockComparisonRepo{createErr: fmt.Errorf("db error")}
	localScan := func(ctx context.Context, rootDir string) ([]domain.FileItem, error) { return nil, nil }
	driveScan := func(ctx context.Context, folderID string) ([]domain.FileItem, error) { return nil, nil }

	uc := NewCompareUseCase(repo, localScan, driveScan)
	_, err := uc.Execute(context.Background(), CompareInput{
		LocalPath: "D:\\Photos",
		DrivePath: "drive-id",
	})

	if err == nil {
		t.Fatal("Execute() should return error when repo create fails")
	}
}

func TestCompareUseCase_RepoSaveFilesError(t *testing.T) {
	repo := &mockComparisonRepo{saveErr: fmt.Errorf("db error")}
	localScan := func(ctx context.Context, rootDir string) ([]domain.FileItem, error) {
		return []domain.FileItem{{RelativePath: "", Name: "a", Extension: ".txt", Size: 1}}, nil
	}
	driveScan := func(ctx context.Context, folderID string) ([]domain.FileItem, error) { return nil, nil }

	uc := NewCompareUseCase(repo, localScan, driveScan)
	_, err := uc.Execute(context.Background(), CompareInput{
		LocalPath: "D:\\Photos",
		DrivePath: "drive-id",
	})

	if err == nil {
		t.Fatal("Execute() should return error when save files fails")
	}
}

func TestToComparisonFiles(t *testing.T) {
	items := []domain.FileItem{
		{RelativePath: "docs/", Name: "report", Extension: ".pdf", Size: 4096},
		{RelativePath: "", Name: "photo", Extension: ".jpg", Size: 2048},
	}

	files := toComparisonFiles(42, items, domain.FileStatusLocalOnly)

	if len(files) != 2 {
		t.Fatalf("len = %d, want 2", len(files))
	}

	if files[0].ComparisonID != 42 {
		t.Errorf("ComparisonID = %d, want 42", files[0].ComparisonID)
	}
	if files[0].Status != domain.FileStatusLocalOnly {
		t.Errorf("Status = %q, want %q", files[0].Status, domain.FileStatusLocalOnly)
	}
	if files[0].Filename != "report" {
		t.Errorf("Filename = %q, want %q", files[0].Filename, "report")
	}
}
