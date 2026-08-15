# Contributing to term-agent

Thank you for your interest in contributing to `term-agent`.

## Development Workflow

1. **Fork and Clone**: Fork the repository on GitHub and clone your fork locally.
2. **Branch Naming**: Create a topic branch using conventional prefixes:
   - `feat/feature-name`
   - `fix/bug-description`
   - `docs/documentation-update`
   - `style/formatting-change`
3. **Coding Standards**:
   - Follow standard Go idioms and formatting standards (`gofmt -w .`).
   - Run `go vet ./...` before committing.
   - Do not include emojis in documentation, commit messages, or PR titles.
4. **Testing**:
   - Write unit and integration tests for new functionality.
   - Run all tests locally: `go test ./...`
   - Verify race detection: `go test -race ./...`
5. **Submitting Pull Requests**:
   - Open a Pull Request against the `develop` branch.
   - Fill out the PR template completely.
   - Ensure CI checks pass.

## Code Style

- Use standard `gofmt` code formatting.
- Ensure exported functions have clear Go docstrings.
- Keep dependencies minimal and audit open-source licenses before adding third-party packages.
