# Monay Security Policy

**Last Updated:** December 2024  
**Owner:** Lawrence Segal  
**Contact:** levi@segaltechnologies.com

## Overview

Monay is a personal wealth management application for private family use. The application serves a small number of authorized family members with no public network access. This document outlines the security controls and practices in place to protect financial data.

## Scope

This policy applies to:
- All Monay application components (services, database, frontend)
- Development and production environments
- All data received from third-party integrations (Plaid)

## Data Classification

| Classification | Description | Examples |
|----------------|-------------|----------|
| Confidential | Financial account data, credentials | Plaid access tokens, holdings data |
| Internal | Application configuration | API endpoints, service configs |
| Public | Documentation | README, architecture docs |

## Access Control

- **Principle of Least Privilege**: Access granted only as needed
- **Unique Accounts**: No shared credentials; individual accounts for all services
- **Multi-Factor Authentication**: Required for:
  - Cloud provider accounts (AWS)
  - Source code repositories (GitHub)
  - Development machines
- **Network Isolation**: Application accessible only via Tailscale private network

## Data Protection

### Encryption in Transit
- All external API communication uses TLS 1.2 or higher
- Internal service communication secured via Tailscale (WireGuard)
- HTTPS enforced for all web interfaces

### Encryption at Rest
- Database volumes encrypted (filesystem-level encryption)
- Development machines use full-disk encryption (FileVault/LUKS)
- Secrets stored in environment variables, never in source code

### Data Retention
- Financial data retained for personal record-keeping
- Plaid access tokens stored securely and rotated as needed
- No data shared with third parties beyond Plaid integration

## Credential Management

- API credentials stored in environment variables
- Credentials excluded from version control via `.gitignore`
- Production secrets managed separately from development

## Vulnerability Management

- Dependencies regularly updated via `go mod tidy`
- GitHub Dependabot alerts monitored for security vulnerabilities
- Container base images updated periodically

## Incident Response

In the event of a suspected security incident:
1. Immediately revoke affected Plaid access tokens
2. Rotate all API credentials
3. Review access logs for unauthorized activity
4. Notify affected parties if applicable

## Third-Party Integrations

### Plaid
- OAuth-based account linking (no credential storage)
- Read-only access to investment data
- Access tokens stored encrypted, never logged

## Review Schedule

This policy is reviewed annually or when significant changes occur to the application architecture.

---

*This is a private family application with a small number of authorized users (family members only). No public network access. Controls are appropriate for the risk profile of personal financial data management.*

