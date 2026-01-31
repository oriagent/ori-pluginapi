// ori-plugin-new is a CLI tool for scaffolding new Ori Agent plugins.
//
// Usage:
//
//	ori-plugin-new my-plugin-name
//	ori-plugin-new my-plugin-name --author "John Doe" --email "john@example.com"
//
// Install:
//
//	go install github.com/oriagent/ori-pluginapi/cmd/ori-plugin-new@latest
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	authorName  = flag.String("author", "", "Author name for maintainers section")
	authorEmail = flag.String("email", "", "Author email for maintainers section")
	description = flag.String("desc", "", "Plugin description")
	outputDir   = flag.String("output", "", "Output directory (defaults to plugin name)")
	withWebPage = flag.Bool("web", false, "Include web page scaffolding")
	withFiles   = flag.Bool("files", false, "Include file attachment scaffolding")
	force       = flag.Bool("force", false, "Overwrite existing directory")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <plugin-name>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Creates a new Ori Agent plugin with all necessary boilerplate.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s my-awesome-plugin\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s my-plugin --author \"Jane Doe\" --email \"jane@example.com\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s my-plugin --web --files\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: plugin name is required")
		flag.Usage()
		os.Exit(1)
	}

	pluginName := flag.Arg(0)

	// Validate plugin name
	if !isValidPluginName(pluginName) {
		fmt.Fprintf(os.Stderr, "Error: invalid plugin name '%s'\n", pluginName)
		fmt.Fprintln(os.Stderr, "Plugin names should contain only lowercase letters, numbers, and hyphens")
		os.Exit(1)
	}

	// Determine output directory
	outDir := *outputDir
	if outDir == "" {
		outDir = pluginName
	}

	// Check if directory exists
	if _, err := os.Stat(outDir); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "Error: directory '%s' already exists\n", outDir)
		fmt.Fprintln(os.Stderr, "Use --force to overwrite")
		os.Exit(1)
	}

	// Build template data
	data := TemplateData{
		PluginName:       pluginName,
		PluginNameSnake:  toSnakeCase(pluginName),
		PluginNamePascal: toPascalCase(pluginName),
		AuthorName:       getOrDefault(*authorName, "Your Name"),
		AuthorEmail:      getOrDefault(*authorEmail, "you@example.com"),
		Description:      getOrDefault(*description, "A plugin that does amazing things"),
		WithWebPage:      *withWebPage,
		WithFiles:        *withFiles,
	}

	// Create directory structure
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Generate files
	files := []struct {
		name     string
		template string
	}{
		{"plugin.yaml", pluginYAMLTemplate},
		{"main.go", mainGoTemplate},
		{"go.mod", goModTemplate},
		{"Makefile", makefileTemplate},
		{".gitignore", gitignoreTemplate},
		{"CLAUDE.md", claudeMdTemplate},
	}

	for _, f := range files {
		path := filepath.Join(outDir, f.name)
		if err := generateFile(path, f.template, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", f.name, err)
			os.Exit(1)
		}
		fmt.Printf("  Created %s\n", path)
	}

	// Print success message
	fmt.Printf("\n✅ Plugin '%s' created successfully!\n\n", pluginName)
	fmt.Println("Next steps:")
	fmt.Printf("  1. cd %s\n", outDir)
	fmt.Println("  2. Edit plugin.yaml to define your parameters and operations")
	fmt.Println("  3. Implement your handlers in main.go")
	fmt.Println("  4. Run 'make build' to compile")
	fmt.Println("  5. Run 'make deploy' to copy to ori-agent")
	fmt.Println("")
}

type TemplateData struct {
	PluginName       string
	PluginNameSnake  string
	PluginNamePascal string
	AuthorName       string
	AuthorEmail      string
	Description      string
	WithWebPage      bool
	WithFiles        bool
}

func isValidPluginName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	// Can't start with hyphen or number
	if name[0] == '-' || (name[0] >= '0' && name[0] <= '9') {
		return false
	}
	return true
}

func toSnakeCase(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

func toPascalCase(s string) string {
	parts := strings.Split(strings.ReplaceAll(s, "-", "_"), "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func generateFile(path, templateStr string, data TemplateData) error {
	tmpl, err := template.New("file").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

// ============================================================================
// Templates
// ============================================================================

var pluginYAMLTemplate = `name: {{.PluginName}}
version: 0.1.0
description: {{.Description}}
tags: ["utility"]

license: MIT
repository: https://github.com/yourusername/{{.PluginName}}

maintainers:
  - name: {{.AuthorName}}
    email: {{.AuthorEmail}}

platforms:
  - os: darwin
    architectures: [amd64, arm64]
  - os: linux
    architectures: [amd64, arm64]

requirements:
  min_ori_version: "0.0.25"
  api_version: "v1"

# Uncomment and customize if your plugin needs configuration
# config:
#   variables:
#     - key: api_key
#       name: API Key
#       description: Your API key for the service
#       type: password
#       required: true
#     - key: timeout
#       name: Timeout
#       description: Request timeout in seconds
#       type: int
#       required: false
#       default_value: 30

tool_definition:
  description: "{{.Description}}"
  parameters:
    - name: operation
      type: string
      description: "The operation to perform"
      required: true
      enum: [status, list, create]

    - name: name
      type: string
      description: "Name parameter for create operation"
      required: false

  operations:
    status:
      parameters: []

    list:
      parameters: []

    create:
      parameters:
        - name: name
          type: string
          description: "Name of the item to create"
          required: true
{{if .WithFiles}}
accepts_files:
  extensions: [txt, json, csv]
  file_operations: [create]
{{end}}{{if .WithWebPage}}
web_pages:
  - dashboard
{{end}}
`

var mainGoTemplate = `package main

// To regenerate: make generate
// Or run directly: ori-plugin-gen -yaml=plugin.yaml -output={{.PluginNameSnake}}_generated.go

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/oriagent/ori-pluginapi"
)

//go:embed plugin.yaml
var configYAML string

// {{.PluginNamePascal}}Tool implements the PluginTool interface
type {{.PluginNamePascal}}Tool struct {
	pluginapi.BasePlugin
}

// ============================================================================
// Operation Handlers
// ============================================================================
// Handlers follow the naming convention: handle{OperationPascalCase}
// The code generator creates a registry that maps operations to these handlers.

func handleStatus(ctx context.Context, t *{{.PluginNamePascal}}Tool, params *Params) (string, error) {
	return "Plugin is running!", nil
}

func handleList(ctx context.Context, t *{{.PluginNamePascal}}Tool, params *Params) (string, error) {
	// Example: Return a table result
	result := pluginapi.NewTableResult(
		"Items",
		[]string{"Name", "Status"},
		[]map[string]string{
			{"Name": "item-1", "Status": "active"},
			{"Name": "item-2", "Status": "pending"},
		},
	)
	result.Description = "Found 2 items"
	return result.ToJSON()
}

func handleCreate(ctx context.Context, t *{{.PluginNamePascal}}Tool, params *Params) (string, error) {
	if params.Name == "" {
		return "", fmt.Errorf("name is required for create operation")
	}
	return fmt.Sprintf("Created item: %s", params.Name), nil
}
{{if .WithFiles}}
// ============================================================================
// File Attachment Handler
// ============================================================================

func handleCreateWithFiles(ctx context.Context, t *{{.PluginNamePascal}}Tool, params *Params, files []pluginapi.FileAttachment) (string, error) {
	if len(files) == 0 {
		return handleCreate(ctx, t, params)
	}

	var results []string
	for _, f := range files {
		results = append(results, fmt.Sprintf("Processed file: %s (%d bytes)", f.Name, f.Size))
	}
	return fmt.Sprintf("Created %s with %d files:\n%s", params.Name, len(files), strings.Join(results, "\n")), nil
}
{{end}}{{if .WithWebPage}}
// ============================================================================
// Web Page Handlers
// ============================================================================

func serveDashboardPage(t *{{.PluginNamePascal}}Tool, query map[string]string) (string, string, error) {
	html := ` + "`" + `<!DOCTYPE html>
<html>
<head>
    <title>{{.PluginName}} Dashboard</title>
    <style>
        body { font-family: system-ui; padding: 2rem; background: #1a1a1a; color: #e5e5e5; }
        h1 { color: #6366f1; }
        .card { background: #252525; padding: 1rem; border-radius: 8px; margin: 1rem 0; }
    </style>
</head>
<body>
    <h1>{{.PluginName}} Dashboard</h1>
    <div class="card">
        <h2>Status</h2>
        <p>Plugin is running!</p>
    </div>
</body>
</html>` + "`" + `
	return html, "text/html", nil
}
{{end}}
// ============================================================================
// Main
// ============================================================================

func main() {
	pluginapi.ServeGRPCPlugin(&{{.PluginNamePascal}}Tool{}, configYAML)
}
`

var goModTemplate = `module github.com/yourusername/{{.PluginName}}

go 1.25

require github.com/oriagent/ori-pluginapi v0.0.1

// For local development, uncomment and adjust the path:
// replace github.com/oriagent/ori-pluginapi => ../../ori-pluginapi
`

var makefileTemplate = `.PHONY: generate build test deploy clean help

# Extract plugin name from plugin.yaml
PLUGIN_NAME := $(shell grep "^name:" plugin.yaml | awk '{print $$2}')
PLUGIN_NAME_SNAKE := $(subst -,_,$(PLUGIN_NAME))

# Adjust these paths to your environment
ORI_PLUGINS := ../../ori-agent/uploaded_plugins
ORI_TEST := ../../ori-test/uploads
# Try to find ori-plugin-gen: installed globally, or in common local paths
ORI_PLUGIN_GEN := $(shell which ori-plugin-gen 2>/dev/null || \
	(test -f ../ori-pluginapi/bin/ori-plugin-gen && echo "../ori-pluginapi/bin/ori-plugin-gen") || \
	(test -f ../../ori-pluginapi/bin/ori-plugin-gen && echo "../../ori-pluginapi/bin/ori-plugin-gen") || \
	echo "ori-plugin-gen")

# Default target
all: build

# Generate code from plugin.yaml
generate:
	@echo "Generating code from plugin.yaml..."
	@if [ ! -f "$(ORI_PLUGIN_GEN)" ] && ! which ori-plugin-gen > /dev/null 2>&1; then \
		echo "Installing ori-plugin-gen..." && \
		go install github.com/oriagent/ori-pluginapi/cmd/ori-plugin-gen@latest; \
	fi
	$(ORI_PLUGIN_GEN) -yaml=plugin.yaml -output=$(PLUGIN_NAME_SNAKE)_generated.go

# Build the plugin binary
build: generate
	@echo "Building $(PLUGIN_NAME)..."
	GOWORK=off CGO_ENABLED=0 go build -o $(PLUGIN_NAME) .
	@echo "✅ Built: $(PLUGIN_NAME)"

# Run tests
test:
	GOWORK=off go test -v ./...

# Deploy to ori-agent
deploy: build
	@mkdir -p $(ORI_PLUGINS)
	cp $(PLUGIN_NAME) $(ORI_PLUGINS)/
	@echo "✅ Deployed to $(ORI_PLUGINS)/$(PLUGIN_NAME)"

# Deploy to ori-test for testing
test-deploy: build
	@mkdir -p $(ORI_TEST)
	cp $(PLUGIN_NAME) $(ORI_TEST)/
	@echo "✅ Deployed to $(ORI_TEST)/$(PLUGIN_NAME)"

# Clean build artifacts
clean:
	rm -f $(PLUGIN_NAME)
	rm -f *_generated.go
	@echo "✅ Cleaned"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build       - Generate code and build the plugin"
	@echo "  make test        - Run tests"
	@echo "  make deploy      - Build and copy to ori-agent"
	@echo "  make test-deploy - Build and copy to ori-test"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make generate    - Generate code from plugin.yaml"
`

var gitignoreTemplate = `# Build artifacts
{{.PluginName}}
*_generated.go

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work
go.work.sum

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Build directories
bin/
dist/
`

var claudeMdTemplate = `# CLAUDE.md

This file provides guidance to Claude Code when working with the {{.PluginName}} plugin.

## Project Overview

**{{.PluginName}}** is a plugin for the Ori Agent framework. {{.Description}}

## Build & Development

` + "```" + `bash
# Build the plugin
make build

# Run tests
make test

# Deploy to ori-agent
make deploy

# Clean build artifacts
make clean
` + "```" + `

## Architecture

### Plugin Type
Direct gRPC plugin (no go-plugin handshake).

### Core Files
- ` + "`main.go`" + ` - Plugin entry point and operation handlers
- ` + "`{{.PluginNameSnake}}_generated.go`" + ` - Auto-generated from plugin.yaml (DO NOT EDIT)
- ` + "`plugin.yaml`" + ` - Single source of truth for tool definition
- ` + "`Makefile`" + ` - Build commands

### Operations

| Operation | Description |
|-----------|-------------|
| ` + "`status`" + ` | Check plugin status |
| ` + "`list`" + ` | List items |
| ` + "`create`" + ` | Create a new item |

### Handler Pattern

Handlers follow the naming convention ` + "`handle{OperationPascalCase}`" + `:

` + "```" + `go
func handleCreate(ctx context.Context, t *{{.PluginNamePascal}}Tool, params *Params) (string, error) {
    // Implementation
}
` + "```" + `

## Plugin API Reference

Key interfaces used:
- ` + "`pluginapi.PluginTool`" + ` - Core interface (Definition, Call)
- ` + "`pluginapi.BasePlugin`" + ` - Embedded for default implementations

### Structured Results

Return rich UI data:

` + "```" + `go
// Table result
result := pluginapi.NewTableResult("Title", []string{"Col1", "Col2"}, rows)
return result.ToJSON()

// Text result
return "Simple text response", nil
` + "```" + `

### Settings API

Store and retrieve plugin configuration:

` + "```" + `go
sm := t.Settings()
value, _ := sm.GetString("key")
sm.Set("key", "value")
` + "```" + `

## Code Generation

After modifying ` + "`plugin.yaml`" + `:

` + "```" + `bash
make build  # Regenerates code and builds
` + "```" + `

## Testing

` + "```" + `bash
# Run unit tests
make test

# Test with ori-agent
make deploy
# Restart ori-agent, then test via chat
` + "```" + `
`
