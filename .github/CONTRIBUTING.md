# Contributing to User Input MCP

Thank you for considering contributing to User Input MCP! This document outlines the process for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by its Code of Conduct. Please report any unacceptable behavior.

## How Can I Contribute?

### Reporting Bugs

- **Ensure the bug was not already reported** by searching on GitHub under [Issues](https://github.com/nazar256/user-input-mcp/issues).
- If you're unable to find an open issue addressing the problem, [open a new one](https://github.com/nazar256/user-input-mcp/issues/new). Be sure to include a **title and clear description**, as much relevant information as possible, and a **code sample** or an **executable test case** demonstrating the expected behavior that is not occurring.

### Suggesting Enhancements

- Open a new issue with a clear title and detailed description.
- Provide specific examples and step-by-step descriptions of the suggested enhancement.

### Pull Requests

1. Fork the repository and create your branch from `main`.
2. If you've added code, add tests.
3. Ensure the test suite passes.
4. Make sure your code follows the existing style.
5. Issue a pull request!

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/nazar256/user-input-mcp.git
   cd user-input-mcp
   ```

2. Install dependencies:
   - Go 1.18 or higher

3. Build the project:
   ```bash
   go build -o user-input-mcp ./cmd/user-input-mcp
   ```

4. Run tests:
   ```bash
   go test ./...
   ```

## Style Guidelines

- Follow standard Go code conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Write unit tests for new functionality

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License. 