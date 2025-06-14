package tools

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*/.gitignore templates/*/*.* templates/*/src/*
var templatesFS embed.FS

// TemplateData holds the data for template rendering
type TemplateData struct {
	AppName     string
	AppNameSlug string
	Port        int
}

// TemplateFile represents a file to be created from a template
type TemplateFile struct {
	Path    string
	Content []byte
}

// LoadProjectTemplates loads all templates for a given project type
func LoadProjectTemplates(projectType string, data TemplateData) ([]TemplateFile, error) {
	var files []TemplateFile
	
	templateDir := filepath.Join("templates", projectType)
	
	// Walk through all files in the template directory
	err := fs.WalkDir(templatesFS, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if d.IsDir() {
			return nil
		}
		
		// Read the template file
		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}
		
		// Process the template
		tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}
		
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", path, err)
		}
		
		// Calculate the relative path from the template directory
		relPath, err := filepath.Rel(templateDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		
		files = append(files, TemplateFile{
			Path:    relPath,
			Content: buf.Bytes(),
		})
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}
	
	return files, nil
}

// GetAvailableProjectTypes returns a list of available project types
func GetAvailableProjectTypes() ([]string, error) {
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}
	
	var types []string
	for _, entry := range entries {
		if entry.IsDir() {
			types = append(types, entry.Name())
		}
	}
	
	return types, nil
}

// CreateAppNameSlug creates a slug from the app name
func CreateAppNameSlug(appName string) string {
	return strings.ToLower(strings.ReplaceAll(appName, " ", "-"))
}