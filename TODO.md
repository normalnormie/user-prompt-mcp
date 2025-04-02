# User Input MCP Server Plan

## Overview
We're implementing an MCP server in Golang that will allow Cursor to request user input during generation. The server will act as a bridge between Cursor and the user, providing a mechanism for the LLM to ask for additional input without ending its generation.

## Implementation Requirements
1. Create an MCP server using Golang
2. Implement the "user input prompt" tool
3. Create a minimal GUI for user input collection
4. Support stdio transport for communication with Cursor

## Research Summary
- MCP (Model Context Protocol) is an open protocol for connecting LLMs with external tools
- Cursor supports MCP integration through stdio transport
- There are two main Golang MCP libraries:
  - mark3labs/mcp-go (more stars, more recent updates, comprehensive)
  - metoro-io/mcp-golang (simpler API according to some sources)
- We'll use mark3labs/mcp-go for our implementation as it appears to be more mature

## Implementation Plan

### 1. Basic Project Setup
- [x] Initialize Go module
- [x] Add mark3labs/mcp-go dependency
- [x] Create basic directory structure (cmd, internal, pkg)
- [x] Set up Github workflow for testing

### 2. Core MCP Server Implementation
- [x] Implement basic MCP server with stdio transport
- [x] Create "user_prompt" tool definition
- [x] Create handler for user_prompt tool

### 3. GUI Implementation
- [x] Research cross-platform GUI options for Go (Fyne, webview, zenity)
- [x] Implement simple input dialog for user prompts
- [x] Ensure GUI works on both Linux and MacOS

### 4. Testing
- [x] Write unit tests for core functionality
- [x] Create integration tests for MCP server
- [x] Test with actual Cursor integration

### 5. Documentation & Distribution
- [x] Write comprehensive README
- [x] Document configuration options
- [x] Create installation/usage instructions
- [x] Set up GitHub Actions for cross-platform binary distribution
- [x] Create installation scripts for Linux, macOS, and Windows

## Technical Decisions

### MCP Library Selection
We chose mark3labs/mcp-go because:
- More active development
- Better documentation
- Larger community

### GUI Implementation
We implemented the GUI using:
- **Zenity** for Linux: Simple shell utility for GUI dialogs
- **osascript** for macOS: Built-in Apple Script support for dialogs

This approach provides a lightweight solution with minimal dependencies that works across platforms.

## Implementation Details

### User Prompt Tool
We implemented a "user_prompt" tool that can be called by Cursor to request additional input from the user. The tool accepts:
- `prompt`: The message to display to the user
- `title`: The title of the dialog window (optional)

### GUI Implementation
We created a cross-platform GUI implementation that uses:
- Zenity on Linux
- AppleScript on macOS

It displays a simple dialog box with a text input field.

### Timeout Handling
The prompt service includes timeout handling to prevent the MCP server from hanging indefinitely if the user doesn't respond. The default timeout is 20 minutes and can be customized through:
- Command line flag: `--timeout <seconds>`
- Environment variable: `USER_PROMPT_TIMEOUT=<seconds>`

For example:
```bash
# Using command line flag
user-prompt-mcp --timeout 600  # Set timeout to 10 minutes

# Using environment variable
export USER_PROMPT_TIMEOUT=1800  # Set timeout to 30 minutes
user-prompt-mcp
```

### Binary Distribution
We've implemented an automated build and release process that:
1. Cross-compiles binaries for multiple platforms (Linux, macOS, Windows) and architectures (amd64, arm64)
2. Attaches these binaries to GitHub releases
3. Provides installation scripts for easy installation without requiring Go:
   - `install.sh` for Linux and macOS
   - `install.bat` for Windows

This makes the tool accessible to users who don't have Go installed on their systems.

## Conclusion
We've successfully implemented a minimal but functional MCP server for user input during Cursor generation. The implementation is cross-platform, lightweight, and follows the MCP specification.

## Open Questions
- What's the best format for the user prompt UI?
- Should we support different types of prompts (text, multiple choice, etc.)?
- How to handle timeout scenarios if user doesn't respond?
- Should we implement additional MCP features beyond the basic user prompt tool?

## Next Steps
1. Set up initial project structure and implement a basic MCP server
2. Implement user prompt tool with simple GUI
3. Test with Cursor and refine as needed 