# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-04-10

### Added
- Initial release of the User Prompt MCP
- Implemented user_prompt tool for requesting input during generation
- Cross-platform GUI support for Linux and macOS
- Timeout handling for user prompts (default: 20 minutes)
- Configurable timeout via command-line flag (--timeout) and environment variable (USER_PROMPT_TIMEOUT)
- Comprehensive documentation and tests 