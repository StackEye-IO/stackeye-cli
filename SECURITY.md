# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.2.x   | Yes       |
| < 0.2   | No        |

As the CLI is pre-1.0, only the latest minor release receives security updates.

## Reporting a Vulnerability

If you discover a security vulnerability in the StackEye CLI, please report it responsibly. **Do not open a public GitHub issue for security vulnerabilities.**

### How to Report

Send an email to **security@stackeye.io** with:

- A description of the vulnerability
- Steps to reproduce or a proof-of-concept
- The affected version(s)
- Any potential impact assessment

### What to Expect

| Stage | Timeline |
|-------|----------|
| Acknowledgement | Within 48 hours |
| Initial assessment | Within 7 days |
| Fix or mitigation | Within 90 days |

We will work with you to understand the issue and coordinate a fix. If you would like to be credited for the discovery, let us know in your report.

### Scope

The following are in scope:

- Authentication and credential handling
- Command injection via CLI arguments
- Sensitive data exposure (API keys, tokens)
- Path traversal in file operations
- Dependencies with known vulnerabilities

The following are out of scope:

- Issues in third-party dependencies that do not affect the CLI (report these upstream)
- Denial of service via resource exhaustion on local machine
- Social engineering attacks

### Disclosure

We follow coordinated disclosure. We ask that you:

1. Allow us reasonable time to address the issue before public disclosure
2. Make a good faith effort to avoid privacy violations, data destruction, or service disruption
3. Do not access or modify other users' data

We commit to:

1. Acknowledging your report promptly
2. Keeping you informed of our progress
3. Crediting you (if desired) when the fix is released
