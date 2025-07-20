package lc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/layered-flow/layered-code/internal/helpers"
	"github.com/mark3labs/mcp-go/mcp"
)

type LcChromeVisitResult struct {
	AppName     string `json:"app_name"`
	URL         string `json:"url"`
	PageContent string `json:"page_content"`
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
}

func LcChromeVisit(appName, url string) (LcChromeVisitResult, error) {
	result := LcChromeVisitResult{
		AppName: appName,
		URL:     url,
		Success: false,
	}

	if err := helpers.ValidateAppName(appName); err != nil {
		result.Message = fmt.Sprintf("Invalid app name: %v", err)
		return result, fmt.Errorf("invalid app name: %w", err)
	}

	if url == "" {
		result.Message = "URL cannot be empty"
		return result, fmt.Errorf("URL cannot be empty")
	}

	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		result.Message = fmt.Sprintf("Failed to get apps directory: %v", err)
		return result, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := os.MkdirAll(appPath, constants.AppsDirectoryPerms); err != nil {
		result.Message = fmt.Sprintf("Failed to create app directory: %v", err)
		return result, fmt.Errorf("failed to create app directory: %w", err)
	}

	// Create a new Chrome instance
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set a timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pageContent string
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second), // Wait for JavaScript to load
		chromedp.OuterHTML("html", &pageContent),
	)

	if err != nil {
		result.Message = fmt.Sprintf("Failed to visit URL: %v", err)
		return result, fmt.Errorf("failed to visit URL: %w", err)
	}

	result.PageContent = pageContent
	result.Success = true
	result.Message = fmt.Sprintf("Successfully visited %s and retrieved page content", url)

	return result, nil
}

func LcChromeVisitCli() error {
	if len(os.Args) < 5 {
		return fmt.Errorf("usage: %s tool lc_chrome_visit <app_name> <url>", os.Args[0])
	}

	appName := os.Args[3]
	url := os.Args[4]

	result, err := LcChromeVisit(appName, url)
	if err != nil {
		return err
	}

	log.Printf("Chrome Visit Result: %s", result.Message)
	if result.Success {
		log.Printf("Page content length: %d characters", len(result.PageContent))
	}

	return nil
}

func LcChromeVisitMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
		URL     string `json:"url"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := LcChromeVisit(args.AppName, args.URL)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}