package domain

import "time"

type Comparison struct {
	ID             uint
	LocalPath      string
	DrivePath      string
	StartedAt      time.Time
	FinishedAt     *time.Time
	TotalLocalOnly int
	TotalDriveOnly int
	TotalBoth      int
}

type FileStatus string

const (
	FileStatusLocalOnly FileStatus = "local_only"
	FileStatusDriveOnly FileStatus = "drive_only"
	FileStatusBoth      FileStatus = "both"
)

type ComparisonFile struct {
	ID           uint
	ComparisonID uint
	Filename     string
	Extension    string
	RelativePath string
	Size         int64
	Status       FileStatus
}

type IgnorePattern struct {
	ID      uint
	Pattern string
}

type FileItem struct {
	RelativePath string
	Name         string
	Extension    string
	Size         int64
}
