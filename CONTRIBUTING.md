# Contributing to Imagine

First off, thank you for considering contributing to Imagine! It's people like you that make Imagine such a great tool.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct:
- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples**
- **Describe the behavior you observed and expected**
- **Include logs and error messages**
- **Note your Go version and OS**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description of the proposed functionality**
- **Include examples of how the feature would be used**
- **Explain why this enhancement would be useful**

### Pull Requests

1. Fork the repo and create your branch from `master`
2. If you've added code that should be tested, add tests
3. Ensure the test suite passes (`make test`)
4. Make sure your code follows the existing style
5. Write a clear commit message

## Development Setup

1. Install Go 1.19 or later
2. Install libvips (see README for platform-specific instructions)
3. Clone your fork:
   ```bash
   git clone https://github.com/your-username/imagine.git
   cd imagine
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Run tests:
   ```bash
   make test
   ```

## Project Structure

```
imagine/
├── imagine.go           # Core library
├── store*.go           # Storage backends
├── hasher.go           # Hashing utilities
├── cmd/                # CLI application
├── src/server/         # Standalone server
└── *_test.go          # Test files
```

## Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting PR
- Include integration tests for new features
- Test with different image formats and sizes

## Style Guide

- Follow standard Go conventions
- Use `gofmt` to format your code
- Keep functions small and focused
- Add comments for exported functions
- Use meaningful variable names

## Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters
- Reference issues and pull requests after the first line

Example:
```
Add WebP format conversion support

- Implement WebP encoding/decoding
- Add quality parameter for WebP
- Update documentation

Fixes #123
```

## Questions?

Feel free to open an issue with your question or reach out through GitHub Discussions.

Thank you for contributing! 🎉