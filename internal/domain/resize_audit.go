package domain

import "strings"

type ResizeAuditResult struct {
	Pending  []FileItem // source files without a matching resized file
	Resized  []FileItem // source files that have a matching resized file
	ExtraRes []FileItem // resized files without a matching source file
}

func ResizeAudit(sourceFiles []FileItem, resizedFiles []FileItem, suffix string) *ResizeAuditResult {
	resizedIndex := make(map[string]struct{}, len(resizedFiles))
	for _, f := range resizedFiles {
		key := resizeKey(f.RelativePath, f.Name, f.Extension)
		resizedIndex[key] = struct{}{}
	}

	sourceIndex := make(map[string]struct{}, len(sourceFiles))
	for _, f := range sourceFiles {
		key := resizeKey(f.RelativePath, f.Name, f.Extension)
		sourceIndex[key] = struct{}{}
	}

	result := &ResizeAuditResult{}

	for _, f := range sourceFiles {
		expectedKey := resizeKey(f.RelativePath, f.Name+suffix, f.Extension)
		if _, found := resizedIndex[expectedKey]; found {
			result.Resized = append(result.Resized, f)
		} else {
			result.Pending = append(result.Pending, f)
		}
	}

	for _, f := range resizedFiles {
		baseName := stripSuffix(f.Name, suffix)
		originalKey := resizeKey(f.RelativePath, baseName, f.Extension)
		if _, found := sourceIndex[originalKey]; !found {
			result.ExtraRes = append(result.ExtraRes, f)
		}
	}

	return result
}

func resizeKey(relativePath, name, extension string) string {
	return relativePath + name + extension
}

func stripSuffix(name string, suffix string) string {
	if strings.HasSuffix(name, suffix) {
		return name[:len(name)-len(suffix)]
	}
	return name
}
