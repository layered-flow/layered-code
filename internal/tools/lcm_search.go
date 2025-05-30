package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/tools/git"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// Types
type LcmSearchMatch struct {
	Index         int      `json:"index"`
	Timestamp     string   `json:"timestamp"`
	Commit        string   `json:"commit"`
	Summary       string   `json:"summary"`
	MatchedFields []string `json:"matched_fields"`
	Context       string   `json:"context,omitempty"`
}

type LcmSearchResult struct {
	Success     bool             `json:"success"`
	Matches     []LcmSearchMatch `json:"matches"`
	TotalFound  int              `json:"total_found"`
	MaxExceeded bool             `json:"max_exceeded,omitempty"`
	Message     string           `json:"message,omitempty"`
}

// LcmSearch searches through LayeredChangeMemory entries
func LcmSearch(appName string, pattern string, caseSensitive bool, maxResults int, fieldFilter string) (LcmSearchResult, error) {
	if err := git.ValidateAppName(appName); err != nil {
		return LcmSearchResult{}, err
	}

	if pattern == "" {
		return LcmSearchResult{}, fmt.Errorf("search pattern is required")
	}

	// Default max results if not specified
	if maxResults <= 0 {
		maxResults = 50
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcmSearchResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := git.ValidateAppPath(appPath); err != nil {
		return LcmSearchResult{}, err
	}

	// Check if LCM file exists
	lcmPath := filepath.Join(appPath, ".layered_change_memory.yaml")
	if _, err := os.Stat(lcmPath); os.IsNotExist(err) {
		return LcmSearchResult{
			Success: true,
			Matches: []LcmSearchMatch{},
			Message: "No layered change memory found",
		}, nil
	}

	// Read the YAML file
	data, err := os.ReadFile(lcmPath)
	if err != nil {
		return LcmSearchResult{}, fmt.Errorf("failed to read LCM file: %w", err)
	}

	// Parse YAML
	var entries []git.LayeredChangeMemoryEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return LcmSearchResult{}, fmt.Errorf("failed to parse LCM file: %w", err)
	}

	// Prepare search pattern
	searchPattern := pattern
	if !caseSensitive {
		searchPattern = strings.ToLower(pattern)
	}

	// Validate field filter
	validFields := map[string]bool{
		"":              true, // all fields
		"all":           true,
		"summary":       true,
		"considerations": true,
		"follow_up":     true,
		"commit_message": true,
	}
	if !validFields[fieldFilter] {
		return LcmSearchResult{}, fmt.Errorf("invalid field_filter: %s (must be one of: all, summary, considerations, follow_up, commit_message)", fieldFilter)
	}

	// Search through entries
	var matches []LcmSearchMatch
	for i, entry := range entries {
		var matchedFields []string
		var matchContext string

		// Search in summary
		if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "summary" {
			text := entry.Summary
			searchText := text
			if !caseSensitive {
				searchText = strings.ToLower(text)
			}
			if strings.Contains(searchText, searchPattern) {
				matchedFields = append(matchedFields, "summary")
				if matchContext == "" {
					matchContext = getMatchContext(text, pattern, caseSensitive)
				}
			}
		}

		// Search in commit message
		if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "commit_message" {
			text := entry.CommitMessage
			searchText := text
			if !caseSensitive {
				searchText = strings.ToLower(text)
			}
			if strings.Contains(searchText, searchPattern) {
				matchedFields = append(matchedFields, "commit_message")
				if matchContext == "" {
					matchContext = getMatchContext(text, pattern, caseSensitive)
				}
			}
		}

		// Search in considerations
		if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "considerations" {
			for _, consideration := range entry.Considerations {
				searchText := consideration
				if !caseSensitive {
					searchText = strings.ToLower(consideration)
				}
				if strings.Contains(searchText, searchPattern) {
					matchedFields = append(matchedFields, "considerations")
					if matchContext == "" {
						matchContext = getMatchContext(consideration, pattern, caseSensitive)
					}
					break
				}
			}
		}

		// Search in follow-up
		if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "follow_up" {
			text := entry.FollowUp
			searchText := text
			if !caseSensitive {
				searchText = strings.ToLower(text)
			}
			if strings.Contains(searchText, searchPattern) {
				matchedFields = append(matchedFields, "follow_up")
				if matchContext == "" {
					matchContext = getMatchContext(text, pattern, caseSensitive)
				}
			}
		}

		// Add to matches if found
		if len(matchedFields) > 0 {
			matches = append(matches, LcmSearchMatch{
				Index:         i,
				Timestamp:     entry.Timestamp,
				Commit:        entry.Commit,
				Summary:       entry.Summary,
				MatchedFields: matchedFields,
				Context:       matchContext,
			})

			// Check if we've hit the max results
			if len(matches) >= maxResults {
				// Check if there are more matches
				maxExceeded := false
				for j := i + 1; j < len(entries); j++ {
					if wouldMatch(entries[j], searchPattern, caseSensitive, fieldFilter) {
						maxExceeded = true
						break
					}
				}
				return LcmSearchResult{
					Success:     true,
					Matches:     matches,
					TotalFound:  len(matches),
					MaxExceeded: maxExceeded,
				}, nil
			}
		}
	}

	return LcmSearchResult{
		Success:    true,
		Matches:    matches,
		TotalFound: len(matches),
	}, nil
}

// getMatchContext returns a snippet of text around the match
func getMatchContext(text, pattern string, caseSensitive bool) string {
	searchText := text
	searchPattern := pattern
	if !caseSensitive {
		searchText = strings.ToLower(text)
		searchPattern = strings.ToLower(pattern)
	}

	index := strings.Index(searchText, searchPattern)
	if index == -1 {
		return ""
	}

	// Get context around match (40 chars before, pattern, 40 chars after)
	contextStart := index - 40
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := index + len(pattern) + 40
	if contextEnd > len(text) {
		contextEnd = len(text)
	}

	context := text[contextStart:contextEnd]
	if contextStart > 0 {
		context = "..." + context
	}
	if contextEnd < len(text) {
		context = context + "..."
	}

	return context
}

// wouldMatch checks if an entry would match without adding it to results
func wouldMatch(entry git.LayeredChangeMemoryEntry, searchPattern string, caseSensitive bool, fieldFilter string) bool {
	// Check summary
	if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "summary" {
		text := entry.Summary
		if !caseSensitive {
			text = strings.ToLower(text)
		}
		if strings.Contains(text, searchPattern) {
			return true
		}
	}

	// Check commit message
	if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "commit_message" {
		text := entry.CommitMessage
		if !caseSensitive {
			text = strings.ToLower(text)
		}
		if strings.Contains(text, searchPattern) {
			return true
		}
	}

	// Check considerations
	if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "considerations" {
		for _, consideration := range entry.Considerations {
			text := consideration
			if !caseSensitive {
				text = strings.ToLower(text)
			}
			if strings.Contains(text, searchPattern) {
				return true
			}
		}
	}

	// Check follow-up
	if fieldFilter == "" || fieldFilter == "all" || fieldFilter == "follow_up" {
		text := entry.FollowUp
		if !caseSensitive {
			text = strings.ToLower(text)
		}
		if strings.Contains(text, searchPattern) {
			return true
		}
	}

	return false
}

// CLI
func LcmSearchCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("lcm_search requires at least 2 arguments: app_name pattern\nUsage: layered-code tool lcm_search <app_name> <pattern> [--case-sensitive] [--max-results N] [--field-filter FIELD]")
	}

	appName := args[0]
	pattern := args[1]
	caseSensitive := false
	maxResults := 50
	fieldFilter := ""

	// Parse optional arguments
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--case-sensitive", "-c":
			caseSensitive = true
		case "--max-results", "-m":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxResults)
				i++
			}
		case "--field-filter", "-f":
			if i+1 < len(args) {
				fieldFilter = args[i+1]
				i++
			}
		}
	}

	result, err := LcmSearch(appName, pattern, caseSensitive, maxResults, fieldFilter)
	if err != nil {
		return fmt.Errorf("failed to search LCM entries: %w", err)
	}

	if result.Message != "" {
		fmt.Println(result.Message)
		return nil
	}

	if len(result.Matches) == 0 {
		fmt.Printf("No matches found for pattern: %s\n", pattern)
		return nil
	}

	fmt.Printf("Found %d matches for pattern: %s\n\n", result.TotalFound, pattern)
	
	for _, match := range result.Matches {
		fmt.Printf("[%d] %s | %s\n", match.Index, match.Timestamp, match.Commit)
		fmt.Printf("    Summary: %s\n", match.Summary)
		fmt.Printf("    Matched in: %s\n", strings.Join(match.MatchedFields, ", "))
		if match.Context != "" {
			fmt.Printf("    Context: %s\n", match.Context)
		}
		fmt.Println()
	}

	if result.MaxExceeded {
		fmt.Printf("(Results limited to %d matches. More matches exist.)\n", maxResults)
	}

	return nil
}

// MCP
func LcmSearchMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName       string `json:"app_name"`
		Pattern       string `json:"pattern"`
		CaseSensitive bool   `json:"case_sensitive"`
		MaxResults    int    `json:"max_results"`
		FieldFilter   string `json:"field_filter"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	if args.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	result, err := LcmSearch(args.AppName, args.Pattern, args.CaseSensitive, args.MaxResults, args.FieldFilter)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}