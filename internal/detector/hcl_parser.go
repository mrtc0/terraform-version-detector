package detector

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// parseRequiredVersion parses HCL content and extracts the required_version
func parseRequiredVersion(content []byte) (string, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(content, "terraform.tf")
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}

	// Extract terraform block
	bodyContent, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: terraformBlockName,
			},
		},
	})
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to parse body: %s", diags.Error())
	}

	// Find terraform block
	if len(bodyContent.Blocks) == 0 {
		return "", fmt.Errorf("no terraform block found")
	}

	terraformBlock := bodyContent.Blocks[0]
	
	// Use PartialContent to extract only the required_version attribute
	// This allows other blocks (like required_providers) to exist
	terraformContent, _, diags := terraformBlock.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: requiredVersionAttr},
		},
	})
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to parse terraform block: %s", diags.Error())
	}

	// Find required_version attribute
	requiredVersionAttrVal, found := terraformContent.Attributes[requiredVersionAttr]
	if !found {
		return "", fmt.Errorf("required_version not found in terraform block")
	}

	// Get the value as string
	val, diags := requiredVersionAttrVal.Expr.Value(nil)
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to evaluate required_version: %s", diags.Error())
	}

	constraint := val.AsString()
	return extractVersionFromConstraint(constraint)
}

// extractVersionFromConstraint extracts a concrete version from a constraint string
func extractVersionFromConstraint(constraint string) (string, error) {
	constraint = strings.TrimSpace(constraint)
	if constraint == "" {
		return "", fmt.Errorf("empty constraint")
	}

	// Remove 'v' prefix if present
	constraint = strings.TrimPrefix(constraint, "v")

	// Pattern to match version numbers (e.g., 1.5.0, 1.0.0)
	versionRegex := regexp.MustCompile(`\d+\.\d+\.\d+`)

	// Extract the first version number found
	matches := versionRegex.FindString(constraint)
	if matches == "" {
		return "", fmt.Errorf("no version found in constraint: %s", constraint)
	}

	return matches, nil
}
