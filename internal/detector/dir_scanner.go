package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findRequiredVersionInDir scans .tf files in the directory and finds required_version
// Returns error if multiple required_version declarations are found
func findRequiredVersionInDir(path string) (string, error) {
	// Read directory entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var foundVersions []versionFound
	
	// Scan all .tf files in the directory (not subdirectories)
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Only process .tf files
		if !strings.HasSuffix(entry.Name(), tfFileExtension) {
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		// Try to parse required_version
		version, err := parseRequiredVersion(content)
		if err == nil {
			// Found required_version
			foundVersions = append(foundVersions, versionFound{
				file:    entry.Name(),
				version: version,
			})
		}
	}

	// Check results
	if len(foundVersions) == 0 {
		return "", fmt.Errorf("no required_version found in .tf files")
	}

	if len(foundVersions) > 1 {
		fileNames := make([]string, len(foundVersions))
		for i, v := range foundVersions {
			fileNames[i] = v.file
		}
		return "", fmt.Errorf("multiple required_version found in files: %s", 
			strings.Join(fileNames, ", "))
	}

	return foundVersions[0].version, nil
}

// versionFound represents a found version in a file
type versionFound struct {
	file    string
	version string
}
