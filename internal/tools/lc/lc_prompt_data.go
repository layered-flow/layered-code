package lc

import (
	_ "embed"
)

// Prompt represents a single prompt with metadata
type Prompt struct {
	ID          int    `json:"id"`
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content,omitempty"`
}

// PromptListResult represents the result of listing prompts
type PromptListResult struct {
	Prompts []Prompt `json:"prompts"`
	Count   int      `json:"count"`
}

// PromptReadResult represents the result of reading a prompt
type PromptReadResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// Embedded prompt files
//go:embed data/general_principles.md
var generalPrinciplesPrompt string

// PromptKey represents a composite key for prompt ID and version
type PromptKey struct {
	ID      int
	Version int
}

// promptsMap stores all available prompts with composite keys (ID, Version)
var promptsMap = map[PromptKey]Prompt{
	{ID: 1, Version: 1}: {
		ID:          1,
		Version:     1,
		Name:        "General Principles",
		Description: "Collaborative coding assistant principles for adaptive, safe, and empowering development",
		Content:     generalPrinciplesPrompt,
	},
	{ID: 1, Version: 2}: {
		ID:          1,
		Version:     2,
		Name:        "General Principles",
		Description: "Enhanced collaborative coding assistant principles with improved safety and user empowerment",
		Content:     generalPrinciplesPrompt + "\n\n## Version 2 Updates\n- Enhanced safety protocols\n- Improved user empowerment features",
	},
}