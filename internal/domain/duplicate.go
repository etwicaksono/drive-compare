package domain

import (
	"fmt"
	"regexp"
	"sort"
)

type DuplicatePattern struct {
	ID    string
	Label string
	Regex *regexp.Regexp
}

type DuplicateGroup struct {
	BaseName     string
	Extension    string
	RelativePath string
	Files        []FileItem
}

func BuiltinPatterns() []DuplicatePattern {
	return []DuplicatePattern{
		{ID: "windows_copy", Label: "Windows copy (N)", Regex: regexp.MustCompile(`^(.+)\s\(\d+\)$`)},
		{ID: "macos_copy", Label: "macOS copy", Regex: regexp.MustCompile(`^(.+)\scopy(?:\s\d+)?$`)},
		{ID: "underscore_num", Label: "Underscore _N", Regex: regexp.MustCompile(`^(.+)_\d{1,3}$`)},
		{ID: "dash_num", Label: "Dash -N", Regex: regexp.MustCompile(`^(.+)-\d{1,3}$`)},
	}
}

func GetPatternsByIDs(ids []string) []DuplicatePattern {
	all := BuiltinPatterns()
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	var result []DuplicatePattern
	for _, p := range all {
		if idSet[p.ID] {
			result = append(result, p)
		}
	}
	return result
}

func ExtractBaseName(name string, patterns []DuplicatePattern) string {
	for _, p := range patterns {
		matches := p.Regex.FindStringSubmatch(name)
		if len(matches) >= 2 {
			return matches[1]
		}
	}
	return name
}

func DetectDuplicates(files []FileItem, patterns []DuplicatePattern) []DuplicateGroup {
	type groupKey struct {
		baseName     string
		extension    string
		relativePath string
		size         int64
	}

	groups := make(map[groupKey][]FileItem)
	order := make([]groupKey, 0)

	for _, f := range files {
		base := ExtractBaseName(f.Name, patterns)
		key := groupKey{baseName: base, extension: f.Extension, relativePath: f.RelativePath, size: f.Size}
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], f)
	}

	var result []DuplicateGroup
	for _, key := range order {
		files := groups[key]
		if len(files) > 1 {
			result = append(result, DuplicateGroup{
				BaseName:     key.baseName,
				Extension:    key.extension,
				RelativePath: key.relativePath,
				Files:        files,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return fmt.Sprintf("%s%s%s", result[i].BaseName, result[i].Extension, result[i].RelativePath) <
			fmt.Sprintf("%s%s%s", result[j].BaseName, result[j].Extension, result[j].RelativePath)
	})

	return result
}
