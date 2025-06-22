package lc

import (
	_ "embed"
)

// Prompt represents a single prompt with metadata
type Prompt struct {
	ID          string `json:"id"`
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

// promptsMap stores all available prompts
var promptsMap = map[string]Prompt{
	"general_principles": {
		ID:          "general_principles",
		Name:        "General Principles",
		Description: "Collaborative coding assistant principles for adaptive, safe, and empowering development",
		Content:     generalPrinciplesPrompt,
	},
}