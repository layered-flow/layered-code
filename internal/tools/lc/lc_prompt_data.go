package lc

import (
	_ "embed"
)

// Prompt represents a single prompt with metadata
type Prompt struct {
	ID          int    `json:"id"`
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

// promptsMap stores all available prompts with numeric IDs as keys
var promptsMap = map[int]Prompt{
	1: {
		ID:          1,
		Name:        "General Principles",
		Description: "Collaborative coding assistant principles for adaptive, safe, and empowering development",
		Content:     generalPrinciplesPrompt,
	},
}