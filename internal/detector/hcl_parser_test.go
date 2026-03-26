package detector

import (
	"testing"
)

func TestParseRequiredVersion(t *testing.T) {
	tests := []struct {
		name        string
		hclContent  string
		expected    string
		expectError bool
	}{
		{
			name: "simple exact version",
			hclContent: `
terraform {
  required_version = "1.5.0"
}
`,
			expected:    "1.5.0",
			expectError: false,
		},
		{
			name: "version with >= constraint",
			hclContent: `
terraform {
  required_version = ">= 1.5.0"
}
`,
			expected:    "1.5.0",
			expectError: false,
		},
		{
			name: "version with ~> constraint",
			hclContent: `
terraform {
  required_version = "~> 1.5.0"
}
`,
			expected:    "1.5.0",
			expectError: false,
		},
		{
			name: "version with = constraint",
			hclContent: `
terraform {
  required_version = "= 1.5.0"
}
`,
			expected:    "1.5.0",
			expectError: false,
		},
		{
			name: "version with multiple constraints",
			hclContent: `
terraform {
  required_version = ">= 1.0.0, < 2.0.0"
}
`,
			expected:    "1.0.0",
			expectError: false,
		},
		{
			name: "terraform block with required_providers",
			hclContent: `
terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
`,
			expected:    "1.5.0",
			expectError: false,
		},
		{
			name: "no terraform block",
			hclContent: `
resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}
`,
			expected:    "",
			expectError: true,
		},
		{
			name: "terraform block without required_version",
			hclContent: `
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}
`,
			expected:    "",
			expectError: true,
		},
		{
			name: "version with v prefix",
			hclContent: `
terraform {
  required_version = "v1.5.0"
}
`,
			expected:    "1.5.0",
			expectError: false,
		},
		{
			name: "required_version in provider block should not be detected",
			hclContent: `
provider "aws" {
  required_version = "1.5.0"
  region = "us-east-1"
}
`,
			expected:    "",
			expectError: true,
		},
		{
			name: "required_version in resource block should not be detected",
			hclContent: `
resource "aws_instance" "example" {
  required_version = "1.5.0"
  ami = "ami-123456"
}
`,
			expected:    "",
			expectError: true,
		},
		{
			name: "required_version at top level should not be detected",
			hclContent: `
required_version = "1.5.0"

resource "aws_instance" "example" {
  ami = "ami-123456"
}
`,
			expected:    "",
			expectError: true,
		},
		{
			name: "required_version in module block should not be detected",
			hclContent: `
module "vpc" {
  source = "./vpc"
  required_version = "1.5.0"
}
`,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseRequiredVersion([]byte(tt.hclContent))

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if version != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, version)
			}
		})
	}
}
