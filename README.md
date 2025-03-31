# User Input MCP

A Model Context Protocol (MCP) server for Cursor that enables requesting user input during generation. This is mostly AI-generated code.

## Overview

This project implements an MCP server that allows Cursor (or any MCP-compatible client) to request additional input from users during model generation without ending the generation process. It serves as a bridge between the AI model and the user, creating a more interactive experience.

## Features

- **User Input Prompting**: Allows the AI to ask for more information during generation
- **Simple GUI**: Presents input prompts in a dialog box with text wrapping
- **Cross-Platform**: Should work on both Linux (tested) and macOS
- **Stdio Transport**: Integration with Cursor via stdio

## Installation

### Prerequisites

- Go 1.18 or higher
- For GUI functionality:
  - Linux: zenity
  - macOS: osascript (not tested)

### Install from source

   ```bash
   go install github.com/nazar256/user-input-mcp/cmd/user-input-mcp@latest
   ```

## Usage with Cursor

1. Open Cursor and go to Settings
2. Navigate to the MCP Servers section and add a new server by editing JSON:
```json
{
  "mcpServers": {
    "prompt": {
      "command": "user-input-mcp"
    }
  }
}
```

3. Save and refresh

4. Add to global rules something like: "Important - do not finish responding until explicitly told to do so, instead always use MCP `input_user_prompt` and wait for response in all cases, either you feel stuck, you have question or you finished work on a prompt - always communicate with the user using this MCP."

Now when using Cursor, the AI can request additional input from you without ending its generation.

## License

MIT

## Acknowledgements

- [Model Context Protocol](https://modelcontextprotocol.io)
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) 