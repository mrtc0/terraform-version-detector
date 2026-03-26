package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Source represents the source of version detection
type Source string

const (
	// SourceTerraformVersionFile indicates version was found in .terraform-version file
	SourceTerraformVersionFile Source = "terraform-version-file"
	// SourceRequiredVersion indicates version was found in terraform block's required_version
	SourceRequiredVersion Source = "required-version"
	// SourceFallback indicates fallback version was used
	SourceFallback Source = "fallback"
	// SourceNotFound indicates version was not found
	SourceNotFound Source = "not-found"
)

const (
	// tfFileExtension is the Terraform file extension
	tfFileExtension = ".tf"
	// terraformBlockName is the name of the terraform configuration block
	terraformBlockName = "terraform"
	// requiredVersionAttr is the name of the required_version attribute
	requiredVersionAttr = "required_version"
)

// Result represents the result of version detection
type Result struct {
	Version string
	Source  Source
	Error   error
}

// newResult creates a new successful Result
func newResult(version string, source Source) Result {
	return Result{
		Version: version,
		Source:  source,
		Error:   nil,
	}
}

// newErrorResult creates a new error Result
func newErrorResult(err error) Result {
	return Result{
		Version: "",
		Source:  SourceNotFound,
		Error:   err,
	}
}

// DetectVersion detects the Terraform version from various sources
func DetectVersion(path, versionFile, fallbackVersion string) Result {
	// Try .terraform-version file first
	versionFilePath := filepath.Join(path, versionFile)
	if version, err := readVersionFile(versionFilePath); err == nil {
		return newResult(version, SourceTerraformVersionFile)
	}

	// Try parsing required_version from terraform block
	if version, err := findRequiredVersionInDir(path); err == nil {
		return newResult(version, SourceRequiredVersion)
	}

	// Use fallback if provided
	if fallbackVersion != "" {
		return newResult(normalizeVersion(fallbackVersion), SourceFallback)
	}

	return newErrorResult(fmt.Errorf("terraform version not found"))
}

// readVersionFile reads and parses the .terraform-version file
func readVersionFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(content))
	if version == "" {
		return "", fmt.Errorf("version file is empty")
	}

	return normalizeVersion(version), nil
}

// normalizeVersion normalizes version string by removing 'v' prefix and trimming whitespace
func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	return version
}
