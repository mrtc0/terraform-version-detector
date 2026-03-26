package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mrtc0/terraform-version-detector/internal/detector"
)

const (
	// Environment variable names
	envInputPath            = "INPUT_PATH"
	envInputVersionFile     = "INPUT_VERSION-FILE"
	envInputFallbackVersion = "INPUT_FALLBACK-VERSION"
	envGitHubOutput         = "GITHUB_OUTPUT"

	// Default values
	defaultPath        = "."
	defaultVersionFile = ".terraform-version"
	stdoutPath         = "/dev/stdout"
)

func main() {
	// Get inputs from environment variables
	path, versionFile, fallback := getInputs()

	// Get output file path from GitHub Actions
	outputFile := os.Getenv(envGitHubOutput)
	if outputFile == "" {
		// For local testing, use stdout
		outputFile = stdoutPath
	}

	// Run detection
	exitCode := run(path, versionFile, fallback, outputFile)
	os.Exit(exitCode)
}

// getInputs reads input parameters from environment variables
func getInputs() (path, versionFile, fallback string) {
	path = os.Getenv(envInputPath)
	if path == "" {
		path = defaultPath
	}

	versionFile = os.Getenv(envInputVersionFile)
	if versionFile == "" {
		versionFile = defaultVersionFile
	}

	fallback = os.Getenv(envInputFallbackVersion)

	return path, versionFile, fallback
}

// run executes the version detection and writes output
func run(path, versionFile, fallback, outputFile string) int {
	// Detect version
	result := detector.DetectVersion(path, versionFile, fallback)

	// Check for errors
	if result.Error != nil {
		log.Printf("Error: %v", result.Error)
		return 1
	}

	// Log success
	log.Printf("Detected Terraform version: %s (source: %s)", result.Version, result.Source)

	// Write output
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
