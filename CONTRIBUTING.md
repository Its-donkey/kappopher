# Contributing to Twitch Bot

Thank you for your interest in contributing to this project! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and constructive in all interactions. We welcome contributors of all experience levels.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- A Twitch developer account (for testing API calls)

### Setting Up the Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/twitch-bot.git
   cd twitch-bot
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/jeffkabot/twitch-bot.git
   ```
4. Install dependencies:
   ```bash
   go mod tidy
   ```

## How to Contribute

### Reporting Bugs

Before submitting a bug report:
- Check the existing issues to avoid duplicates
- Collect relevant information (Go version, OS, error messages)

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md) when creating an issue.

### Suggesting Features

We welcome feature suggestions! Please use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:
- A clear description of the feature
- The use case and benefits
- Any implementation ideas you have

### Submitting Code Changes

1. Create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the coding standards below

3. Write or update tests as needed

4. Run the tests:
   ```bash
   go test ./...
   ```

5. Ensure the code builds:
   ```bash
   go build ./...
   ```

6. Commit your changes with a clear message:
   ```bash
   git commit -m "Add feature: description of your changes"
   ```

7. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

8. Open a Pull Request against the `main` branch

## Coding Standards

### Go Style

- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
- Run `go fmt` on all code before committing
- Run `go vet` to catch common issues
- Use meaningful variable and function names
- Add comments for exported functions and types

### API Implementation Guidelines

When adding new Twitch API endpoints:

1. Create the endpoint function in the appropriate `helix/*.go` file
2. Define request/response structs with proper JSON tags
3. Include validation for required parameters
4. Handle pagination where applicable
5. Add documentation in `docs/` following the existing format:
   - Include required OAuth scopes
   - Provide Go code examples
   - Include sample JSON responses

### Testing

- Write unit tests for new functionality
- Test edge cases and error conditions
- Mock HTTP responses for API tests

### Documentation

- Update relevant documentation when changing functionality
- Follow the existing documentation format in `docs/`
- Include code examples with proper error handling

## Pull Request Process

1. Ensure all tests pass
2. Update documentation as needed
3. Fill out the PR template completely
4. Link any related issues
5. Request review from maintainers

PRs will be reviewed for:
- Code quality and style
- Test coverage
- Documentation completeness
- Compatibility with existing code

## Questions?

If you have questions about contributing, feel free to open an issue with the "question" label.

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.
