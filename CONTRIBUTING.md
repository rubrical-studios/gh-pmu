# Contributing to gh-pmu

Thank you for your interest in contributing to gh-pmu!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/gh-pmu.git`
3. Create a branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit with a descriptive message
7. Push and open a Pull Request

## Development Setup

See [docs/development.md](docs/development.md) for:
- Prerequisites
- Building from source
- Running tests
- Project structure

## Code Guidelines

### Style

- Follow standard Go conventions
- Run `gofmt -w .` before committing
- Keep functions focused and small
- Add comments for exported functions

### Commits

Use conventional commit format:

```
type: short description

Longer explanation if needed.

Fixes #123
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Tests

- Add tests for new functionality
- Maintain or improve coverage
- Tests should be deterministic

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...
```

## Pull Request Process

1. **Update documentation** if adding/changing commands
2. **Add tests** for new functionality
3. **Keep PRs focused** - one feature or fix per PR
4. **Link issues** - Reference related issues in description
5. **Wait for review** - Maintainers will review and provide feedback

### PR Description Template

```markdown
## Summary
Brief description of changes.

## Changes
- Change 1
- Change 2

## Testing
How was this tested?

## Related Issues
Fixes #123
```

## Reporting Issues

### Bug Reports

Include:
- gh-pmu version (`gh pmu --version`)
- Go version (`go version`)
- OS and version
- Steps to reproduce
- Expected vs actual behavior
- Error messages (if any)

### Feature Requests

Include:
- Use case description
- Proposed solution (if any)
- Alternatives considered

## Questions?

- Open a [Discussion](https://github.com/rubrical-studios/gh-pmu/discussions)
- Check existing [Issues](https://github.com/rubrical-studios/gh-pmu/issues)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
