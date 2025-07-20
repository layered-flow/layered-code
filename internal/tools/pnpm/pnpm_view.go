package pnpm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type PnpmViewResult struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	Homepage         string            `json:"homepage"`
	License          string            `json:"license,omitempty"`
	Repository       interface{}       `json:"repository,omitempty"`
	Keywords         []string          `json:"keywords,omitempty"`
	Maintainers      []interface{}     `json:"maintainers,omitempty"`
	Dependencies     map[string]string `json:"dependencies,omitempty"`
	DevDependencies  map[string]string `json:"devDependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies,omitempty"`
}

// PnpmView retrieves package information from npm registry
func PnpmView(packageName string) (PnpmViewResult, error) {
	// Trim whitespace
	packageName = strings.TrimSpace(packageName)
	
	// Validate package name
	if packageName == "" {
		return PnpmViewResult{}, fmt.Errorf("package name is required")
	}

	// Split package name and version if @ is present (but not at the start for scoped packages)
	basePackage := packageName
	if strings.HasPrefix(packageName, "@") {
		// Handle scoped packages like @scope/package@version
		slashIdx := strings.Index(packageName, "/")
		if slashIdx == -1 {
			return PnpmViewResult{}, fmt.Errorf("invalid scoped package format, expected @scope/package")
		}
		
		// Check if there's a version after the package name
		afterSlash := packageName[slashIdx+1:]
		if afterSlash == "" {
			return PnpmViewResult{}, fmt.Errorf("invalid scoped package format, expected @scope/package")
		}
		
		// Extract base package name for validation
		versionIdx := strings.LastIndex(packageName, "@")
		if versionIdx > slashIdx {
			basePackage = packageName[:versionIdx]
		}
		
		// Validate scoped package format
		parts := strings.Split(basePackage, "/")
		if len(parts) != 2 || len(parts[0]) <= 1 || parts[1] == "" {
			return PnpmViewResult{}, fmt.Errorf("invalid scoped package format, expected @scope/package")
		}
	} else {
		// Handle non-scoped packages like package@version
		atIdx := strings.Index(packageName, "@")
		if atIdx > 0 {
			basePackage = packageName[:atIdx]
		}
	}

	// Basic validation to prevent command injection
	if strings.ContainsAny(packageName, ";|&`$(){}[]<>") {
		return PnpmViewResult{}, fmt.Errorf("invalid characters in package name")
	}

	// Validate that npm is available (works with both npm and pnpm)
	if _, err := exec.LookPath("npm"); err != nil {
		return PnpmViewResult{}, fmt.Errorf("npm command not found: %w", err)
	}

	// Build command - using npm view since it works with both npm and pnpm
	cmd := exec.Command("npm", "view", packageName, "--json")
	
	// Capture output
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		errStr := errBuf.String()
		
		// Check for common npm errors and provide cleaner messages
		if strings.Contains(errStr, "E404") || strings.Contains(errStr, "404") {
			return PnpmViewResult{}, fmt.Errorf("Package '%s' not found in npm registry", packageName)
		}
		
		if strings.Contains(errStr, "EINVALIDPACKAGENAME") || strings.Contains(errStr, "Invalid package name") {
			return PnpmViewResult{}, fmt.Errorf("invalid package name '%s'", packageName)
		}
		
		// For other errors, return a clean error without exposing raw npm output
		return PnpmViewResult{}, fmt.Errorf("failed to retrieve package information for '%s'", packageName)
	}

	// Parse JSON output
	var npmData struct {
		Name             string            `json:"name"`
		Version          string            `json:"version"`
		Description      string            `json:"description"`
		Homepage         string            `json:"homepage"`
		License          interface{}       `json:"license"`
		Repository       interface{}       `json:"repository"`
		Keywords         []string          `json:"keywords"`
		Maintainers      []interface{}     `json:"maintainers"`
		Dependencies     map[string]string `json:"dependencies"`
		DevDependencies  map[string]string `json:"devDependencies"`
		PeerDependencies map[string]string `json:"peerDependencies"`
	}

	if err := json.Unmarshal(outBuf.Bytes(), &npmData); err != nil {
		return PnpmViewResult{}, fmt.Errorf("failed to parse npm view output: %w", err)
	}

	// Extract license string (npm can return string or object)
	var licenseStr string
	switch v := npmData.License.(type) {
	case string:
		licenseStr = v
	case map[string]interface{}:
		if typeVal, ok := v["type"].(string); ok {
			licenseStr = typeVal
		}
	}

	return PnpmViewResult{
		Name:             npmData.Name,
		Version:          npmData.Version,
		Description:      npmData.Description,
		Homepage:         npmData.Homepage,
		License:          licenseStr,
		Repository:       npmData.Repository,
		Keywords:         npmData.Keywords,
		Maintainers:      npmData.Maintainers,
		Dependencies:     npmData.Dependencies,
		DevDependencies:  npmData.DevDependencies,
		PeerDependencies: npmData.PeerDependencies,
	}, nil
}

// CLI
func PnpmViewCli() error {
	args := os.Args[3:]

	if len(args) != 1 {
		return fmt.Errorf("usage: layered-code tool pnpm_view <package_name[@version]>")
	}

	packageName := args[0]
	
	result, err := PnpmView(packageName)
	if err != nil {
		return err
	}

	// Display result
	fmt.Printf("Package: %s@%s\n", result.Name, result.Version)
	if result.Description != "" {
		fmt.Printf("Description: %s\n", result.Description)
	}
	if result.License != "" {
		fmt.Printf("License: %s\n", result.License)
	}
	if result.Homepage != "" {
		fmt.Printf("Homepage: %s\n", result.Homepage)
	}
	
	// Display repository
	if result.Repository != nil {
		switch repo := result.Repository.(type) {
		case string:
			fmt.Printf("Repository: %s\n", repo)
		case map[string]interface{}:
			if url, ok := repo["url"].(string); ok {
				fmt.Printf("Repository: %s\n", url)
			}
		}
	}
	
	// Display keywords
	if len(result.Keywords) > 0 {
		fmt.Printf("Keywords: %s\n", strings.Join(result.Keywords, ", "))
	}
	
	// Display maintainers
	if len(result.Maintainers) > 0 {
		fmt.Println("\nMaintainers:")
		for _, m := range result.Maintainers {
			if maintainer, ok := m.(map[string]interface{}); ok {
				name := maintainer["name"]
				email := maintainer["email"]
				if name != nil {
					fmt.Printf("  - %s", name)
					if email != nil {
						fmt.Printf(" <%s>", email)
					}
					fmt.Println()
				}
			}
		}
	}
	
	// Display dependencies
	if len(result.Dependencies) > 0 {
		fmt.Println("\nDependencies:")
		for dep, version := range result.Dependencies {
			fmt.Printf("  %s: %s\n", dep, version)
		}
	}
	
	// Display dev dependencies
	if len(result.DevDependencies) > 0 {
		fmt.Println("\nDev Dependencies:")
		for dep, version := range result.DevDependencies {
			fmt.Printf("  %s: %s\n", dep, version)
		}
	}
	
	// Display peer dependencies
	if len(result.PeerDependencies) > 0 {
		fmt.Println("\nPeer Dependencies:")
		for dep, version := range result.PeerDependencies {
			fmt.Printf("  %s: %s\n", dep, version)
		}
	}
	
	if len(result.Dependencies) == 0 && len(result.DevDependencies) == 0 && len(result.PeerDependencies) == 0 {
		fmt.Println("\nNo dependencies")
	}

	return nil
}

// MCP
func PnpmViewMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		PackageName string `json:"package_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := PnpmView(args.PackageName)
	if err != nil {
		return nil, err
	}

	// Convert result to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}