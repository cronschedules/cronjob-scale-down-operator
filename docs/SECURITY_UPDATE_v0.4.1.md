# Security Update v0.4.1

This release addresses several security vulnerabilities identified in the project dependencies.

## Vulnerabilities Fixed

### High Severity
- **CVE-2025-22868** - `golang.org/x/oauth2`
  - **Impact**: Unexpected memory consumption during token parsing
  - **Fixed**: Updated from v0.23.0 to v0.30.0
  - **CVSS Score**: 7.5

### Medium Severity
- **CVE-2025-22872** - `golang.org/x/net`
  - **Fixed**: Updated from v0.30.0 to v0.43.0
  
- **CVE-2025-22870** - `golang.org/x/net`
  - **Fixed**: Updated from v0.30.0 to v0.43.0

### Unknown Severity
- **CVE-2025-47907** - Go stdlib
  - **Fixed**: Updated Go version from 1.23.0 to 1.23.12

## Additional Improvements

### Error Handling
- Improved graceful handling of missing target resources
- Target resources not found now log as INFO instead of ERROR with stack traces
- Cleaner production logs

### Dependency Updates
- Updated all golang.org/x packages to latest secure versions
- Improved overall security posture

## Verification

All tests pass and the project builds successfully with the updated dependencies:

```bash
# Run tests
make test

# Build project
make build

# Check for vulnerabilities
go list -m golang.org/x/oauth2 golang.org/x/net
```

## Upgrade Instructions

1. Pull the latest code
2. Rebuild your container images using the new version
3. Deploy using the updated image tag `0.4.1`

For Helm deployments:
```bash
helm upgrade cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator --version 0.4.1
```

## Container Images

Updated secure images are available at:
- `ghcr.io/cronschedules/cronjob-scale-down-operator:0.4.1`
- `cronschedules/cronjob-scale-down-operator:0.4.1`
