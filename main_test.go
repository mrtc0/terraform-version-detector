package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper functions

func assertExitCode(t *testing.T, got, expected int) {
	t.Helper()
	if got != expected {
		t.Errorf("expected exit code %d, got %d", expected, got)
	}
}

func assertOutputContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("output should contain %q, got: %s", expected, output)
	}
}

func readOutputFile(t *testing.T, path string) string {
	t.Helper()
	output, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	return string(output)
}

func createTerraformVersionFile(t *testing.T, dir, version string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, ".terraform-version"), []byte(version), 0644)
	if err != nil {
		t.Fatalf("failed to create .terraform-version file: %v", err)
	}
}

func createVersionsTF(t *testing.T, dir, constraint string) {
	t.Helper()
	versionsContent := `
terraform {
  required_version = "` + constraint + `"
}
`
	err := os.WriteFile(filepath.Join(dir, "versions.tf"), []byte(versionsContent), 0644)
	if err != nil {
		t.Fatalf("failed to create versions.tf: %v", err)
	}
}

func TestRun(t *testing.T) {
	t.Run("successful detection from .terraform-version", func(t *testing.T) {
		tmpDir := t.TempDir()
		createTerraformVersionFile(t, tmpDir, "1.5.0")
		outputFile := filepath.Join(tmpDir, "output.txt")

		exitCode := run(tmpDir, ".terraform-version", "", outputFile)

		assertExitCode(t, exitCode, 0)
		output := readOutputFile(t, outputFile)
		assertOutputContains(t, output, "version=1.5.0")
		assertOutputContains(t, output, "source=terraform-version-file")
	})

	t.Run("successful detection from required_version", func(t *testing.T) {
		tmpDir := t.TempDir()
		createVersionsTF(t, tmpDir, ">= 1.5.0")
		outputFile := filepath.Join(tmpDir, "output.txt")

		exitCode := run(tmpDir, ".terraform-version", "", outputFile)

		assertExitCode(t, exitCode, 0)
		output := readOutputFile(t, outputFile)
		assertOutputContains(t, output, "version=1.5.0")
		assertOutputContains(t, output, "source=required-version")
	})

	t.Run("use fallback when no version found", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.txt")

		exitCode := run(tmpDir, ".terraform-version", "1.0.0", outputFile)

		assertExitCode(t, exitCode, 0)
		output := readOutputFile(t, outputFile)
		assertOutputContains(t, output, "version=1.0.0")
		assertOutputContains(t, output, "source=fallback")
	})

	t.Run("error when no version found and no fallback", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.txt")

		exitCode := run(tmpDir, ".terraform-version", "", outputFile)

		assertExitCode(t, exitCode, 1)
	})
}

func TestGetInputs(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		expectedPath string
		expectedFile string
		expectedFB   string
	}{
		{
			name: "all inputs provided",
			envVars: map[string]string{
				"INPUT_PATH":             "/custom/path",
				"INPUT_VERSION-FILE":     ".tfversion",
				"INPUT_FALLBACK-VERSION": "1.2.3",
			},
			expectedPath: "/custom/path",
			expectedFile: ".tfversion",
			expectedFB:   "1.2.3",
		},
		{
			name:         "use defaults when not provided",
			envVars:      map[string]string{},
			expectedPath: ".",
			expectedFile: ".terraform-version",
			expectedFB:   "",
		},
		{
			name: "partial inputs",
			envVars: map[string]string{
				"INPUT_PATH": "/my/path",
			},
			expectedPath: "/my/path",
			expectedFile: ".terraform-version",
			expectedFB:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			path, versionFile, fallback := getInputs()

			if path != tt.expectedPath {
				t.Errorf("expected path %q, got %q", tt.expectedPath, path)
			}
			if versionFile != tt.expectedFile {
				t.Errorf("expected versionFile %q, got %q", tt.expectedFile, versionFile)
			}
			if fallback != tt.expectedFB {
				t.Errorf("expected fallback %q, got %q", tt.expectedFB, fallback)
			}
		})
	}
}
