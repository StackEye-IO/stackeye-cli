# Changelog

All notable changes to StackEye CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Dynamic shell completion for context commands (Task #7163)
- Error message formatting utilities (Task #7164)
- Dynamic shell completion for probe IDs (Task #7162)
- Table formatters for billing commands with `-o table` output (Task #7946)
- Upgrade command for self-updating CLI binary (Task #7115)
- Update notifier with background version checking (Task #7114)
- Probe logs command with follow mode (Task #7112)
- Probe import command (Task #7111)
- Probe export command (Task #7110)
- Probe watch command for live status updates (Task #7109)
- `stackeye env` command to list environment variables (Task #7133)
- Markdown docs generation (Task #7132)
- Man page generation using cobra/doc (Task #7131)
- Graceful signal handling for SIGINT/SIGTERM (Task #7130)
- `--timeout` flag and `STACKEYE_TIMEOUT` env var support (Task #7129)
- CONTRIBUTING.md guide (Task #7150)
- RELEASING.md documentation (Task #7156)
- Dependabot configuration for Go modules and Actions (Task #7154)

### Changed

- Refactored output package to use atomic.Value for thread-safe globals (Task #8199)
- Refactored telemetry flush into signal handler cleanup chain (Task #8210)

### Fixed

- Use GetReleaseByTag for specific version downloads (Task #7115)
- ProgressBar integration tests and stderr pipe detection (Task #8198)
- Excessive timeout warning for typo detection (Task #8209)

## [0.2.0-rc.2] - 2026-02-01

### Added

- Scoop manifest support for Windows package management (Task #7138)

## [0.2.0-rc.1] - 2026-02-01

### Added

- Wasabi S3 distribution for CLI binaries
- Homebrew tap push in release workflow (may fail silently)

### Changed

- Release workflow uses AWS CLI for Wasabi upload instead of GoReleaser blobs
- Release workflow switched to GitHub-hosted runner

## [0.1.0-alpha.2] - 2026-02-01

### Added

- Homebrew tap integration for macOS package management (Task #7137)
- API key auth support for CI/CD automation (Task #7948)
- Setup wizard command with unit tests
- Probe dependencies add command

## [0.1.0-alpha.1] - 2026-01-31

Initial alpha release of StackEye CLI.

### Added

- Core CLI framework using Cobra
- Authentication with `stackeye login` and `stackeye logout`
- Context management with `stackeye context`
- Probe management commands: list, create, delete, enable, disable
- Probe link-channel command for associating notification channels
- Probe deps commands for hierarchical alerting
- Interactive probe wizard command
- Interactive setup wizard command
- Alert management commands: list, ack, resolve
- Notification channel commands: list, create, delete
- Status page commands: list, create, update, delete
- Team management commands: list, add-member, remove-member
- Billing commands: status, invoices (with --download), portal
- Table formatters for consistent output across all commands
- JSON and YAML output formats with `-o json` and `-o yaml`
- Configuration file support (`~/.stackeye/config.yaml`)
- Multiple output formats: table, JSON, YAML
- Bash, Zsh, Fish, and PowerShell completion scripts
- Curl/bash install script for quick installation
- GoReleaser configuration for automated releases

[Unreleased]: https://github.com/StackEye-IO/stackeye-cli/compare/v0.2.0-rc.2...HEAD
[0.2.0-rc.2]: https://github.com/StackEye-IO/stackeye-cli/compare/v0.2.0-rc.1...v0.2.0-rc.2
[0.2.0-rc.1]: https://github.com/StackEye-IO/stackeye-cli/compare/v0.1.0-alpha.2...v0.2.0-rc.1
[0.1.0-alpha.2]: https://github.com/StackEye-IO/stackeye-cli/compare/v0.1.0-alpha.1...v0.1.0-alpha.2
[0.1.0-alpha.1]: https://github.com/StackEye-IO/stackeye-cli/releases/tag/v0.1.0-alpha.1
