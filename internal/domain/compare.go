package domain

import "fmt"

type ComparisonResult struct {
	LocalOnly []FileItem
	DriveOnly []FileItem
	Both      []FileItem
}

func CompareFiles(localFiles []FileItem, driveFiles []FileItem) *ComparisonResult {
	driveIndex := make(map[string]struct{}, len(driveFiles))
	for _, f := range driveFiles {
		driveIndex[fileKey(f)] = struct{}{}
	}

	localIndex := make(map[string]struct{}, len(localFiles))
	for _, f := range localFiles {
		localIndex[fileKey(f)] = struct{}{}
	}

	result := &ComparisonResult{}

	for _, f := range localFiles {
		key := fileKey(f)
		if _, found := driveIndex[key]; found {
			result.Both = append(result.Both, f)
		} else {
			result.LocalOnly = append(result.LocalOnly, f)
		}
	}

	for _, f := range driveFiles {
		key := fileKey(f)
		if _, found := localIndex[key]; !found {
			result.DriveOnly = append(result.DriveOnly, f)
		}
	}

	return result
}

func fileKey(f FileItem) string {
	return fmt.Sprintf("%s%s%s|%d", f.RelativePath, f.Name, f.Extension, f.Size)
}
