# Chart Testing Setup

This directory contains the configuration and setup for automated Helm chart testing using [chart-testing](https://github.com/helm/chart-testing) and GitHub Actions.

## Overview

The chart testing setup includes:

- **Lint Testing**: Validates chart structure, templates, and values
- **Install Testing**: Tests actual installation in a Kubernetes cluster
- **Security Scanning**: Runs security checks on generated manifests
- **Multi-Values Testing**: Tests chart with different configuration scenarios

## Directory Structure

```
charts/
├── cronjob-scale-down-operator/        # Main Helm chart
│   └── ci/                             # CI test values
│       ├── default-values.yaml         # Standard configuration test
│       ├── minimal-values.yaml         # Minimal resource test
│       ├── ha-values.yaml             # High availability test
│       └── testing-values.yaml        # CI-optimized test values
├── ci/                                 # Legacy CI config (deprecated)
└── index.yaml                          # Chart repository index
.ct.yaml                                # Chart testing configuration
.yamllint                              # YAML linting rules
.github/workflows/chart-test.yml        # GitHub Actions workflow
scripts/test-chart.sh                  # Local testing script
```

## Configuration Files

### `.ct.yaml`
Main configuration for chart-testing tool:
- Specifies chart directories to test
- Configures Helm repositories for dependencies
- Sets testing parameters and validation rules

### `.yamllint`
YAML linting configuration:
- Enforces consistent YAML formatting
- Sets line length and indentation rules
- Configures comment and document formatting

### CI Test Values Files
Multiple values files test different scenarios:

- **`testing-values.yaml`**: Optimized for CI with reduced resources and disabled WebUI
- **`default-values.yaml`**: Standard production-like configuration
- **`minimal-values.yaml`**: Minimal resource allocation for resource-constrained environments
- **`ha-values.yaml`**: High availability configuration with multiple replicas

## GitHub Actions Workflow

The `.github/workflows/chart-test.yml` workflow runs automatically on:
- Push to `main` branch (when chart files change)
- Pull requests affecting chart files

### Workflow Jobs

1. **lint-test**: 
   - Validates chart structure and templates
   - Tests installation in a kind cluster
   - Runs with all CI values files

2. **security-scan**:
   - Scans generated Kubernetes manifests for security issues
   - Uses Checkov for security policy validation

3. **release** (main branch only):
   - Automatically releases charts to GitHub Pages
   - Updates chart repository index

## Local Testing

### Prerequisites

Install required tools:

```bash
# Install Helm
brew install helm

# Install chart-testing
brew install chart-testing

# Install kind (for cluster testing)
brew install kind

# Install kubeval (optional, for manifest validation)
brew install kubeval
```

### Run Local Tests

Use the provided script for comprehensive local testing:

```bash
./scripts/test-chart.sh
```

Or run individual commands:

```bash
# Add required repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Run chart linting
ct lint --target-branch main --chart-dirs charts

# Test template rendering
helm template test-release charts/cronjob-scale-down-operator/ \
  --values charts/cronjob-scale-down-operator/ci/testing-values.yaml \
  --dry-run

# Create kind cluster and test installation
kind create cluster --name chart-testing
helm install test-release charts/cronjob-scale-down-operator/ \
  --values charts/cronjob-scale-down-operator/ci/testing-values.yaml \
  --wait
```

## Test Scenarios

### Standard Installation Test
Tests the chart with default production-like settings:
- Full resource allocation
- WebUI enabled
- All features activated

### Minimal Resource Test
Tests chart with minimal resource requirements:
- Reduced CPU and memory limits
- Optimized for resource-constrained environments

### CI/Testing Environment
Tests chart optimized for CI pipelines:
- WebUI disabled to reduce resource usage
- Extended probe timeouts for slower CI environments
- Leader election disabled for single replica testing

### High Availability Test
Tests chart configured for production HA:
- Multiple replicas
- Anti-affinity rules
- Enhanced resource allocation

## Troubleshooting

### Common Issues

**Chart lint failures:**
- Check YAML syntax and indentation
- Validate template expressions
- Ensure required values are properly defaulted

**Template rendering errors:**
- Verify all referenced values exist
- Check conditional logic in templates
- Test with minimal values files

**Installation failures:**
- Check RBAC permissions
- Verify CRD installation
- Review pod logs and events

### Debugging Commands

```bash
# Debug template rendering
helm template test-release charts/cronjob-scale-down-operator/ \
  --values charts/cronjob-scale-down-operator/ci/testing-values.yaml \
  --debug

# Check rendered manifests
helm get manifest test-release

# View chart dependencies
helm dependency list charts/cronjob-scale-down-operator/

# Validate against Kubernetes API
helm template test-release charts/cronjob-scale-down-operator/ \
  --validate
```

## Contributing

When making changes to the chart:

1. Run local tests with `./scripts/test-chart.sh`
2. Update relevant CI values files if needed
3. Update chart version in `Chart.yaml`
4. Update this documentation if configuration changes

The GitHub Actions workflow will automatically validate your changes and provide feedback on pull requests.

## Additional Resources

- [Helm Chart Testing Guide](https://helm.sh/docs/topics/chart_tests/)
- [Chart Testing Tool Documentation](https://github.com/helm/chart-testing)
- [Kubernetes Best Practices for Charts](https://helm.sh/docs/chart_best_practices/)
- [Security Scanning with Checkov](https://www.checkov.io/)
