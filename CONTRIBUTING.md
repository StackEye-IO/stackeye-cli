# Contributing to StackEye CLI

Thank you for your interest in contributing to the StackEye CLI. This guide covers the development workflow, coding standards, and process for submitting changes.

## Prerequisites

- **Go 1.25** or later
- **Make** (GNU Make)
- **Git**
- **golangci-lint** (installed automatically by Makefile if not present)

## Development Setup

```bash
# Clone the repository
git clone https://github.com/StackEye-IO/stackeye-cli.git
cd stackeye-cli

# Download dependencies
go mod download

# Build the binary
make build

# Run tests
make test

# Run linters
make lint

# Full validation (format check + vet + lint + tests)
make validate
```

The built binary is at `bin/stackeye`.

## Code Style

- Run `make fmt` to format code before committing.
- Run `make lint` to check for lint violations. The project uses [golangci-lint](https://golangci-lint.run/) with the configuration in `.golangci.yml`.
- Run `make vet` to run `go vet`.
- Use `make validate-quick` for a fast pre-commit check (format + vet + lint, no tests).

All CI checks must pass before a pull request can be merged.

## Testing

- Write tests for new functionality. Place tests alongside the code they cover (`foo_test.go` next to `foo.go`).
- End-to-end tests live in `test/e2e/`.
- Run `make test` to execute all tests with the race detector enabled.
- Run `make test-verbose` for verbose output.
- Run `make coverage` to generate a coverage report.

## Making Changes

1. **Fork the repository** and create a feature branch from `main`:
   ```bash
   git checkout -b feat/your-feature main
   ```

2. **Make your changes.** Keep commits focused — one logical change per commit.

3. **Run validation** before pushing:
   ```bash
   make validate
   ```

4. **Push your branch** and open a pull request against `main`.

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `ci`, `chore`.

Examples:
- `feat(probe): add history subcommand`
- `fix(auth): handle expired refresh tokens`
- `docs: update installation instructions`

## Pull Request Process

1. Fill out the PR description with a summary of changes and any relevant context.
2. Ensure CI passes (lint, tests, build).
3. Keep the PR focused. Large changes should be broken into smaller PRs when possible.
4. A maintainer will review your PR. Address feedback and push updates to the same branch.

## Reporting Issues

Open an issue on [GitHub Issues](https://github.com/StackEye-IO/stackeye-cli/issues) with:
- A clear title and description
- Steps to reproduce (for bugs)
- Expected vs. actual behavior
- CLI version (`stackeye version`) and OS/architecture

## Developer Certificate of Origin (DCO)

By contributing to this project, you certify that:

- The contribution was created in whole or in part by you and you have the right to submit it under the MIT license; or
- The contribution is based upon previous work that, to the best of your knowledge, is covered under an appropriate open source license and you have the right to submit that work with modifications under the MIT license.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
