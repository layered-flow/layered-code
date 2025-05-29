# LayeredChangeMemory Example

When making a git commit using the layered-code MCP tool, you MUST include LayeredChangeMemory metadata. This metadata will be stored in a `.layered_change_memory.yaml` file in your app directory to track commit context and decisions for future reference.

## Example Usage

### Via MCP (Model Context Protocol)

When calling the `git_commit` tool via MCP, the `layered_change_memory` parameter is REQUIRED:

```json
{
  "app_name": "my-app",
  "message": "Update homepage hero section copy",
  "layered_change_memory": {
    "summary": "Rewrote hero section text to clarify value proposition and appeal to new users.",
    "considerations": [
      "Kept original hero image as a placeholder—needs update.",
      "Skipped mobile responsiveness for hero section due to time constraints.",
      "Did not replace the title with 'Hello World' as the user did not like that title."
    ],
    "follow_up": "Replace hero image and ensure hero section is mobile-friendly."
  }
}
```

### Example with User Rejections

```json
{
  "app_name": "TestApp1",
  "message": "Update heading text to 'Your no one' and change from H1 to H3",
  "layered_change_memory": {
    "summary": "Changed heading text to 'Your no one' and converted H1 to H3 element",
    "considerations": [
      "User rejected 'Hello World' and 'Hello Everyone' as heading options",
      "Changed semantic HTML from H1 to H3 as specifically requested",
      "Did not add any emojis or decorative elements as user prefers plain text"
    ],
    "follow_up": "Consider reviewing the semantic HTML structure since H3 is now the main page heading"
  }
}
```

### Generated YAML Output

This will create or append to `.layered_change_memory.yaml` in your app directory:

```yaml
- timestamp: "2025-05-29T16:21:54Z"
  commit: "b1e2f7a"
  commit_message: "Update homepage hero section copy"
  summary: "Rewrote hero section text to clarify value proposition and appeal to new users."
  considerations:
    - "Kept original hero image as a placeholder—needs update."
    - "Skipped mobile responsiveness for hero section due to time constraints."
    - "Did not replace the title with 'Hello World' as the user did not like that title."
  follow_up: "Replace hero image and ensure hero section is mobile-friendly."
```

## Key Features

1. **Timestamp**: Automatically generated in UTC format
2. **Commit Hash**: Short 7-character git commit hash
3. **Commit Message**: Limited to 2 lines maximum
4. **Summary**: REQUIRED - One-sentence description of changes
5. **Considerations**: Array of strings (maximum 3) about limitations or decisions
6. **Follow Up**: Optional string for next steps

## Required Fields

When using the MCP interface, the following structure is required:
- `layered_change_memory` (object) - REQUIRED
  - `summary` (string) - REQUIRED - A concise one-sentence description
  - `considerations` (array of strings) - Can be empty array, max 3 items
    - **IMPORTANT**: Include what the user explicitly rejected or didn't want
    - This helps avoid repeating the same mistakes in future iterations
    - Example: "Did not use 'Hello World' as the user rejected that title"
  - `follow_up` (string) - Optional, can be empty string

## CLI Usage

The LayeredChangeMemory feature is only available through the MCP interface. When using the CLI directly, commits will be created without LayeredChangeMemory entries.