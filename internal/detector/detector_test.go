package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectVersion_TerraformVersionFile(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		expectedVer    string
		expectedSource Source
		expectError    bool
	}{
		{
			name:           "simple version",
			fileContent:    "1.5.0",
			expectedVer:    "1.5.0",
			expectedSource: SourceTerraformVersionFile,
			expectError:    false,
		},
		{
			name:           "version with v prefix",
			fileContent:    "v1.5.0",
			expectedVer:    "1.5.0",
			expectedSource: SourceTerraformVersionFile,
			expectError:    false,
		},
		{
			name:           "version with newline",
			fileContent:    "1.5.0\n",
			expectedVer:    "1.5.0",
			expectedSource: SourceTerraformVersionFile,
			expectError:    false,
		},
		{
			name:           "version with spaces",
			fileContent:    "  1.5.0  \n",
			expectedVer:    "1.5.0",
			expectedSource: SourceTerraformVersionFile,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			versionFile := filepath.Join(tmpDir, ".terraform-version")

			// Write test file
			err := os.WriteFile(versionFile, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Test
			result := DetectVersion(tmpDir, ".terraform-version", "")

			if tt.expectError && result.Error == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && result.Error != nil {
				t.Errorf("unexpected error: %v", result.Error)
			}
			if result.Version != tt.expectedVer {
				t.Errorf("expected version %q, got %q", tt.expectedVer, result.Version)
			}
			if result.Source != tt.expectedSource {
				t.Errorf("expected source %q, got %q", tt.expectedSource, result.Source)
			}
		})
	}
}

func TestDetectVersion_NoVersionFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Test without fallback
	result := DetectVersion(tmpDir, ".terraform-version", "")
	if result.Error == nil {
		t.Errorf("expected error when no version file exists")
	}
	if result.Version != "" {
		t.Errorf("expected empty version, got %q", result.Version)
	}
	if result.Source != SourceNotFound {
		t.Errorf("expected source %q, got %q", SourceNotFound, result.Source)
	}
}

func TestDetectVersion_WithFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with fallback
	result := DetectVersion(tmpDir, ".terraform-version", "1.0.0")
	if result.Error != nil {
		t.Errorf("unexpected error with fallback: %v", result.Error)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected fallback version '1.0.0', got %q", result.Version)
	}
	if result.Source != SourceFallback {
		t.Errorf("expected source %q, got %q", SourceFallback, result.Source)
	}
}

func TestDetectVersion_RequiredVersion(t *testing.T) {
	t.Run("from versions.tf", func(t *testing.T) {
		tmpDir := t.TempDir()

		versionsContent := `
terraform {
  required_version = ">= 1.5.0"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "versions.tf"), []byte(versionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create versions.tf: %v", err)
		}

		result := DetectVersion(tmpDir, ".terraform-version", "")
		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}
		if result.Version != "1.5.0" {
			t.Errorf("expected version '1.5.0', got %q", result.Version)
		}
		if result.Source != SourceRequiredVersion {
			t.Errorf("expected source %q, got %q", SourceRequiredVersion, result.Source)
		}
	})

	t.Run("terraform-version file takes precedence over required_version", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .terraform-version file
		err := os.WriteFile(filepath.Join(tmpDir, ".terraform-version"), []byte("1.6.0"), 0644)
		if err != nil {
			t.Fatalf("failed to create .terraform-version: %v", err)
		}

		// Create versions.tf with different version
		versionsContent := `
terraform {
  required_version = ">= 1.5.0"
}
`
		err = os.WriteFile(filepath.Join(tmpDir, "versions.tf"), []byte(versionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create versions.tf: %v", err)
		}

		result := DetectVersion(tmpDir, ".terraform-version", "")
		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}
		// Should use .terraform-version, not required_version
		if result.Version != "1.6.0" {
			t.Errorf("expected version '1.6.0' from .terraform-version, got %q", result.Version)
		}
		if result.Source != SourceTerraformVersionFile {
			t.Errorf("expected source %q, got %q", SourceTerraformVersionFile, result.Source)
		}
	})

	t.Run("required_version before fallback", func(t *testing.T) {
		tmpDir := t.TempDir()

		versionsContent := `
terraform {
  required_version = "~> 1.5.0"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "versions.tf"), []byte(versionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create versions.tf: %v", err)
		}

		// Provide fallback, but required_version should be used
		result := DetectVersion(tmpDir, ".terraform-version", "1.0.0")
		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}
		if result.Version != "1.5.0" {
			t.Errorf("expected version '1.5.0' from required_version, got %q", result.Version)
		}
		if result.Source != SourceRequiredVersion {
			t.Errorf("expected source %q, got %q", SourceRequiredVersion, result.Source)
		}
	})
}
