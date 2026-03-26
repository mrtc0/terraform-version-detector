package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRequiredVersionInDir(t *testing.T) {
	t.Run("single file with required_version", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a versions.tf file
		versionsContent := `
terraform {
  required_version = ">= 1.5.0"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "versions.tf"), []byte(versionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		version, err := findRequiredVersionInDir(tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if version != "1.5.0" {
			t.Errorf("expected version '1.5.0', got %q", version)
		}
	})

	t.Run("multiple tf files, only one has required_version", func(t *testing.T) {
		tmpDir := t.TempDir()

		// versions.tf with required_version
		versionsContent := `
terraform {
  required_version = "~> 1.5.0"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "versions.tf"), []byte(versionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create versions.tf: %v", err)
		}

		// main.tf without required_version
		mainContent := `
resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}
`
		err = os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainContent), 0644)
		if err != nil {
			t.Fatalf("failed to create main.tf: %v", err)
		}

		version, err := findRequiredVersionInDir(tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if version != "1.5.0" {
			t.Errorf("expected version '1.5.0', got %q", version)
		}
	})

	t.Run("multiple files with required_version - should error", func(t *testing.T) {
		tmpDir := t.TempDir()

		// versions.tf with required_version
		versionsContent := `
terraform {
  required_version = ">= 1.5.0"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "versions.tf"), []byte(versionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create versions.tf: %v", err)
		}

		// main.tf also with required_version
		mainContent := `
terraform {
  required_version = ">= 1.4.0"
}
`
		err = os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainContent), 0644)
		if err != nil {
			t.Fatalf("failed to create main.tf: %v", err)
		}

		_, err = findRequiredVersionInDir(tmpDir)
		if err == nil {
			t.Errorf("expected error when multiple required_version found")
		}
		// Check error message contains file names
		errMsg := err.Error()
		if !strings.Contains(errMsg, "multiple") {
			t.Errorf("error message should mention 'multiple': %s", errMsg)
		}
	})

	t.Run("no tf files", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := findRequiredVersionInDir(tmpDir)
		if err == nil {
			t.Errorf("expected error when no .tf files found")
		}
	})

	t.Run("tf files without required_version", func(t *testing.T) {
		tmpDir := t.TempDir()

		mainContent := `
resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainContent), 0644)
		if err != nil {
			t.Fatalf("failed to create main.tf: %v", err)
		}

		_, err = findRequiredVersionInDir(tmpDir)
		if err == nil {
			t.Errorf("expected error when no required_version found")
		}
	})

	t.Run("ignores subdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Root level versions.tf
		versionsContent := `
terraform {
  required_version = ">= 1.5.0"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "versions.tf"), []byte(versionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create versions.tf: %v", err)
		}

		// Create subdirectory with its own versions.tf (should be ignored)
		subDir := filepath.Join(tmpDir, "modules", "app")
		err = os.MkdirAll(subDir, 0755)
		if err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}

		subVersionsContent := `
terraform {
  required_version = ">= 1.6.0"
}
`
		err = os.WriteFile(filepath.Join(subDir, "versions.tf"), []byte(subVersionsContent), 0644)
		if err != nil {
			t.Fatalf("failed to create sub versions.tf: %v", err)
		}

		version, err := findRequiredVersionInDir(tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Should use root level version, not subdirectory
		if version != "1.5.0" {
			t.Errorf("expected version '1.5.0' from root, got %q", version)
		}
	})
}
