package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mrtc0/terraform-version-detector/internal/detector"
)

const (
	envTargetDir       = "TARGET_DIR"
	envVersionFile     = "VERSION_FILE"
	envFallbackVersion = "FALLBACK_VERSION"
	envGitHubOutput    = "GITHUB_OUTPUT" // default output file for GitHub Actions, if not set, will write to stdout

	defaultPath        = "."
	defaultVersionFile = ".terraform-version"

	stdoutPath = "/dev/stdout"
)

func main() {
	path, versionFile, fallback := getInputs()

	outputFile := os.Getenv(envGitHubOutput)
	if outputFile == "" {
		outputFile = stdoutPath
	}

	exitCode := run(path, versionFile, fallback, outputFile)
	os.Exit(exitCode)
}

// getInputs reads input parameters from environment variables
func getInputs() (path, versionFile, fallback string) {
	path = os.Getenv(envTargetDir)
	if path == "" {
		path = defaultPath
	}

	versionFile = os.Getenv(envVersionFile)
	if versionFile == "" {
		versionFile = defaultVersionFile
	}

	fallback = os.Getenv(envFallbackVersion)

	return path, versionFile, fallback
}

// run executes the version detection and writes output
func run(path, versionFile, fallback, outputFile string) int {
	// Detect version
	result := detector.DetectVersion(path, versionFile, fallback)

	if result.Error != nil {
		log.Printf("Error: %v", result.Error)
		return 1
	}

	log.Printf("Detected Terraform version: %s (source: %s)", result.Version, result.Source)

	if err := writeOutputToFile(outputFile, result); err != nil {
		log.Printf("Failed to write output: %v", err)
		return 1
	}

	return 0
}

// writeOutputToFile writes the detection result to a file or stdout
func writeOutputToFile(outputFile string, result detector.Result) error {
	var w io.Writer

	if outputFile != "" && outputFile != stdoutPath {
		f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer f.Close()
		w = f
	} else {
		w = os.Stdout
	}

	return writeOutput(w, result)
}

// writeOutput writes the detection result to an io.Writer
func writeOutput(w io.Writer, result detector.Result) error {
	if _, err := fmt.Fprintf(w, "version=%s\n", result.Version); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}

	if _, err := fmt.Fprintf(w, "source=%s\n", result.Source); err != nil {
		return fmt.Errorf("failed to write source: %w", err)
	}

	return nil
}
