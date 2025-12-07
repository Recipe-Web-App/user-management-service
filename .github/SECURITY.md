# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Do not open a public issue.**

If you discover a security vulnerability in this project, please report it privately via GitHub Security Advisories.

1. Go to the **Security** tab of the repository.
2. Click **Advisories**.
3. Click **New draft security advisory**.

A maintainer will review your report and respond within 48 hours.

## Severity Levels

We categorize issues based on CVSS scores:

- **Critical**: RCE, SQL Injection, Auth Bypass
- **High**: Privilege Escalation, Sensitive Data Exposure
- **Medium**: XSS, CSRF
- **Low**: Information Leakage (non-sensitive)

## Security Features

- **gosec** scanning in CI pipeline
- **CodeQL** analysis
- **Dependabot** security updates

## Disclosure Policy

We follow a Coordinated Disclosure policy. We ask that valid vulnerabilities be kept private until a patch is released
to protect our users.
