package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

type GitShowResult struct {
	Success   bool   `json:"success"`
	IsRepo    bool   `json:"is_repo"`
	Message   string `json:"message,omitempty"`
	Content   string `json:"content,omitempty"`
	Hash      string `json:"hash,omitempty"`
	Author    string `json:"author,omitempty"`
	Date      string `json:"date,omitempty"`
	Subject   string `json:"subject,omitempty"`
	CommitRef string `json:"commit_ref,omitempty"`
}

func GitShow(appName, commitRef string) (GitShowResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitShowResult{}, err
	}

	if err := ValidateAppName(appName); err != nil {
		return GitShowResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitShowResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitShowResult{}, err
	}

	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitShowResult{
			Success: false,
			IsRepo:  false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	if commitRef == "" {
		commitRef = "HEAD"
	}

	args := []string{"show", "--format=fuller", commitRef}
	cmd := exec.Command("git", args...)
	cmd.Dir = appPath

	output, err := cmd.Output()
	if err != nil {
		return GitShowResult{
			Success:   false,
			IsRepo:    true,
			Message:   fmt.Sprintf("Failed to show commit '%s': %s", commitRef, err.Error()),
			CommitRef: commitRef,
		}, nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return GitShowResult{
			Success:   false,
			IsRepo:    true,
			Message:   fmt.Sprintf("Commit '%s' not found", commitRef),
			CommitRef: commitRef,
		}, nil
	}

	hash, author, date, subject := parseCommitInfo(outputStr)

	return GitShowResult{
		Success:   true,
		IsRepo:    true,
		Message:   fmt.Sprintf("Successfully retrieved commit '%s'", commitRef),
		Content:   outputStr,
		Hash:      hash,
		Author:    author,
		Date:      date,
		Subject:   subject,
		CommitRef: commitRef,
	}, nil
}

func parseCommitInfo(output string) (hash, author, date, subject string) {
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "commit ") {
			hash = strings.TrimSpace(strings.TrimPrefix(line, "commit "))
		} else if strings.HasPrefix(line, "Author:") {
			author = strings.TrimSpace(strings.TrimPrefix(line, "Author:"))
		} else if strings.HasPrefix(line, "CommitDate:") {
			date = strings.TrimSpace(strings.TrimPrefix(line, "CommitDate:"))
		} else if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "commit ") && 
			!strings.HasPrefix(line, "Author:") && !strings.HasPrefix(line, "AuthorDate:") && 
			!strings.HasPrefix(line, "Commit:") && !strings.HasPrefix(line, "CommitDate:") && 
			!strings.HasPrefix(line, "diff --git") && !strings.HasPrefix(line, "index ") && 
			!strings.HasPrefix(line, "+++") && !strings.HasPrefix(line, "---") && 
			!strings.HasPrefix(line, "@@") && subject == "" {
			subject = strings.TrimSpace(line)
		}
	}
	
	return hash, author, date, subject
}

func GitShowCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("usage: layered-code git show <app_name> [commit_ref]")
	}

	appName := args[0]
	commitRef := ""
	if len(args) > 1 {
		commitRef = args[1]
	}

	result, err := GitShow(appName, commitRef)
	if err != nil {
		return fmt.Errorf("failed to show commit: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		return nil
	}

	fmt.Print(result.Content)
	return nil
}

func GitShowMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName   string `json:"app_name"`
		CommitRef string `json:"commit_ref,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitShow(args.AppName, args.CommitRef)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}