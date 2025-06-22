package lc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type ComponentExample struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Dependencies []string          `json:"dependencies"`
	Tags         []string          `json:"tags"`
	Files        map[string]string `json:"files"`
}

type LcComponentsParams struct {
	Category string `json:"category,omitempty"`
	Example  string `json:"example,omitempty"`
	Action   string `json:"action"`
	Search   string `json:"search,omitempty"`
}

// Embedded component examples
var componentExamples = map[string][]ComponentExample{
	"layouts": {
		{
			Name:        "admin_panel",
			Description: "Minimal responsive panel layout with collapsible sidebar navigation using @radix-ui/themes",
			Dependencies: []string{
				"@radix-ui/themes",
				"@radix-ui/react-icons",
				"next-themes",
			},
			Tags: []string{
				"layout",
				"responsive",
				"admin",
				"navigation",
				"sidebar",
				"mobile-friendly",
				"radix-ui",
			},
			Files: map[string]string{
				"components/MobileTopBar.tsx": `import { Flex, Button, Box } from "@radix-ui/themes"
import { HamburgerMenuIcon, Cross1Icon } from "@radix-ui/react-icons"
import Icon from "@/components/layout/Icon"

interface MobileTopBarProps {
  isMenuOpen: boolean
  toggleMenu: () => void
}

export default function MobileTopBar({ isMenuOpen, toggleMenu }: MobileTopBarProps) {
  return (
    <Box
      display={{ initial: "block", md: "none" }}
      style={{
        backgroundColor: "var(--gray-2)",
        borderBottom: "1px solid var(--gray-4)",
        position: "sticky",
        top: 0,
        zIndex: 50,
      }}
      p="3"
    >
      <Flex justify="between" align="center" py="1" px="1" pr="2">
        <Icon height={19} />

        <Button
          variant="ghost"
          size="3"
          style={{
            transition: "background-color 0.2s ease-in-out, color 0.2s ease-in-out",
            cursor: "pointer"
          }}
          onClick={toggleMenu}
        >
          {isMenuOpen ? <Cross1Icon height={20} width={20} /> : <HamburgerMenuIcon height={20} width={20} />}
        </Button>
      </Flex>
    </Box>
  )
}`,
				"components/SideMenu.tsx": `import { Flex, Button, Box } from "@radix-ui/themes"
import Logo from "@/components/layout/Logo"

interface MenuItem {
  label: string
  href: string
}

interface SideMenuProps {
  isMenuOpen: boolean
  setIsMenuOpen: (open: boolean) => void
  menuItems: MenuItem[]
}

export default function SideMenu({ isMenuOpen, setIsMenuOpen, menuItems }: SideMenuProps) {
  return (
    <Box
      style={{
        width: "280px",
        height: "100vh",
        backgroundColor: "var(--gray-2)",
        borderRight: "1px solid var(--gray-4)",
        zIndex: 50,
        transition: "transform 0.3s ease-in-out",
        flexShrink: 0,
        transform: isMenuOpen
          ? "translateX(0)"
          : "translateX(-100%)",
      }}
      className={` + "`" + `fixed top-0 left-0 lg:relative lg:!transform-none` + "`" + `}
      p="4"
    >
      <Box mb="3">
        <Logo height={28} />
      </Box>

      <Flex direction="column" gap="2" mt="5">
        {menuItems.map((item) => (
          <Button
            key={item.label}
            variant="ghost"
            color="gray"
            size="2"
            style={{
              justifyContent: "flex-start",
              transition: "background-color 0.2s ease-in-out, color 0.2s ease-in-out",
              cursor: "pointer"
            }}
            onClick={() => setIsMenuOpen(false)}
          >
            {item.label}
          </Button>
        ))}
      </Flex>
    </Box>
  )
}`,
				"components/MobileMenuOverlay.tsx": `import { Box } from "@radix-ui/themes"

interface MobileMenuOverlayProps {
  isMenuOpen: boolean
  setIsMenuOpen: (open: boolean) => void
}

export default function MobileMenuOverlay({ isMenuOpen, setIsMenuOpen }: MobileMenuOverlayProps) {
  if (!isMenuOpen) return null

  return (
    <Box
      display={{ initial: "block", md: "none" }}
      style={{
        position: "fixed",
        inset: 0,
        backgroundColor: "rgba(0, 0, 0, 0.5)",
        zIndex: 40,
      }}
      onClick={() => setIsMenuOpen(false)}
    />
  )
}`,
				"components/MainContent.tsx": `import { Box, Text } from "@radix-ui/themes"

export default function MainContent() {
  return (
    <Box
      style={{
        flexGrow: 1,
      }}
      className="lg:ml-0"
      p="4"
    >
      <Text>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Morbi scelerisque dui at dapibus tincidunt. In eleifend ante sit amet rhoncus cursus. Nullam erat lacus, ultrices vel molestie eu, luctus id enim.</Text>
    </Box>
  )
}`,
				"app/layout.tsx": `import type { Metadata } from "next"

import { ThemeProvider } from 'next-themes'
import { Theme } from "@radix-ui/themes"

import { Geist, Geist_Mono } from "next/font/google"

import "./globals.css"
import "@radix-ui/themes/styles.css"

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
})

export const metadata: Metadata = {
  title: "Create Next App",
  description: "Generated by create next app",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <body className={` + "`" + `${geistSans.variable} ${geistMono.variable} antialiased` + "`" + `}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <Theme
            accentColor="gray"
            grayColor="slate"
            radius="small"
          >
            {children}
          </Theme>
        </ThemeProvider>
      </body>
    </html>
  )
}`,
				"app/page.tsx": `"use client"

import { Flex, Box } from "@radix-ui/themes"
import { useState } from "react"

import MobileTopBar from "@/components/MobileTopBar"
import SideMenu from "@/components/SideMenu"
import MainContent from "@/components/MainContent"
import MobileMenuOverlay from "@/components/MobileMenuOverlay"

export default function Home() {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen)
  }

  const menuItems = [
    { label: "Item", href: "#" }
  ]

  return (
    <Box>
      <MobileTopBar isMenuOpen={isMenuOpen} toggleMenu={toggleMenu} />
      <MobileMenuOverlay isMenuOpen={isMenuOpen} setIsMenuOpen={setIsMenuOpen} />

      <Flex height="100vh">
        <SideMenu
          isMenuOpen={isMenuOpen}
          setIsMenuOpen={setIsMenuOpen}
          menuItems={menuItems}
        />
        <MainContent />
      </Flex>
    </Box>
  )
}`,
			},
		},
	},
}

func LcComponents(params LcComponentsParams) (interface{}, error) {
	switch params.Action {
	case "list":
		return listComponentsData(params.Category)
	case "get":
		if params.Example == "" {
			return nil, fmt.Errorf("example name required for get action")
		}
		if params.Category == "" {
			return nil, fmt.Errorf("category required for get action")
		}
		return getComponentData(params.Category, params.Example)
	case "search":
		if params.Search == "" {
			return nil, fmt.Errorf("search term required")
		}
		return searchComponentsData(params.Search)
	default:
		return nil, fmt.Errorf("invalid action: %s. Use 'list', 'get', or 'search'", params.Action)
	}
}

func listComponentsData(category string) (map[string][]string, error) {
	result := make(map[string][]string)

	if category != "" {
		// List specific category
		if examples, exists := componentExamples[category]; exists {
			names := make([]string, 0, len(examples))
			for _, example := range examples {
				names = append(names, example.Name)
			}
			result[category] = names
		}
	} else {
		// List all categories
		for cat, examples := range componentExamples {
			names := make([]string, 0, len(examples))
			for _, example := range examples {
				names = append(names, example.Name)
			}
			result[cat] = names
		}
	}

	return result, nil
}

func getComponentData(category string, exampleName string) (*ComponentExample, error) {
	examples, exists := componentExamples[category]
	if !exists {
		return nil, fmt.Errorf("category '%s' not found", category)
	}

	for _, example := range examples {
		if example.Name == exampleName {
			return &example, nil
		}
	}

	return nil, fmt.Errorf("example '%s' not found in category '%s'", exampleName, category)
}

func searchComponentsData(searchTerm string) ([]map[string]interface{}, error) {
	searchTerm = strings.ToLower(searchTerm)
	var results []map[string]interface{}

	for category, examples := range componentExamples {
		for _, example := range examples {
			if strings.Contains(strings.ToLower(example.Name), searchTerm) ||
				strings.Contains(strings.ToLower(example.Description), searchTerm) ||
				containsTag(example.Tags, searchTerm) {
				
				results = append(results, map[string]interface{}{
					"category":    category,
					"name":        example.Name,
					"description": example.Description,
					"tags":        example.Tags,
				})
			}
		}
	}

	return results, nil
}

func containsTag(tags []string, search string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), search) {
			return true
		}
	}
	return false
}

// CLI
func LcComponentsCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printLcComponentsHelp()
			return nil
		}
	}

	var params LcComponentsParams
	params.Action = "list" // default action

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--category":
			if i+1 < len(args) {
				params.Category = args[i+1]
				i++
			} else {
				return errors.New("--category requires a value")
			}
		case "--example":
			if i+1 < len(args) {
				params.Example = args[i+1]
				i++
			} else {
				return errors.New("--example requires a value")
			}
		case "--action":
			if i+1 < len(args) {
				params.Action = args[i+1]
				i++
			} else {
				return errors.New("--action requires a value")
			}
		case "--search":
			if i+1 < len(args) {
				params.Action = "search"
				params.Search = args[i+1]
				i++
			} else {
				return errors.New("--search requires a value")
			}
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool lc_components --help' for usage", args[i])
			}
		}
	}

	result, err := LcComponents(params)
	if err != nil {
		return err
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	
	fmt.Println(string(output))
	return nil
}

func printLcComponentsHelp() {
	fmt.Println("Usage: layered-code tool lc_components [options]")
	fmt.Println()
	fmt.Println("Access layered code component examples")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --action <action>    Action to perform: list, get, search (default: list)")
	fmt.Println("  --category <name>    Component category (layouts, forms, features)")
	fmt.Println("  --example <name>     Example name (required for 'get' action)")
	fmt.Println("  --search <term>      Search term (automatically sets action to 'search')")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List all available components")
	fmt.Println("  layered-code tool lc_components")
	fmt.Println()
	fmt.Println("  # List components in a specific category")
	fmt.Println("  layered-code tool lc_components --category layouts")
	fmt.Println()
	fmt.Println("  # Get a specific component example")
	fmt.Println("  layered-code tool lc_components --action get --category layouts --example admin_panel")
	fmt.Println()
	fmt.Println("  # Search for components")
	fmt.Println("  layered-code tool lc_components --search responsive")
}

// MCP
func LcComponentsMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LcComponentsParams

	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	// Default action is list
	if params.Action == "" {
		params.Action = "list"
	}

	result, err := LcComponents(params)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}