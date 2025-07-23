# Charts Repository Migration Guide

## Overview

The CronJob Scale Down Operator Helm charts have been migrated from this repository to a dedicated charts repository for better management, hosting, and CI/CD practices.

## Migration Details

### Repository Information

| Detail | Previous | New |
|--------|----------|-----|
| **Repository** | `z4ck404/cronjob-scale-down-operator` | `cronschedules/charts` |
| **Location** | `/charts/cronjob-scale-down-operator/` | `/cronjob-scale-down-operator/` |
| **Helm Repo URL** | N/A (local only) | `https://cronschedules.github.io/charts` |
| **Chart Structure** | Nested under `/charts` | Root level chart |
| **Migration Date** | July 2025 | - |

### Key Changes

1. **Dedicated Repository**: Charts now have their own repository with proper Helm hosting
2. **GitHub Pages Hosting**: Charts are properly hosted and indexed
3. **Automated CI/CD**: Full testing pipeline with security scanning
4. **Better Documentation**: Comprehensive chart documentation and examples
5. **Multiple Configurations**: Pre-configured values for different scenarios

## Migration Steps

### For New Users

Simply use the new repository:

```bash
# Add the new charts repository
helm repo add cronschedules https://cronschedules.github.io/charts
helm repo update

# Install the operator
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator
```

### For Existing Users

If you were using the old chart location, follow these steps:

#### 1. Backup Current Configuration

```bash
# Get current values
helm get values cronjob-scale-down-operator > my-values.yaml

# Get current status
helm status cronjob-scale-down-operator
```

#### 2. Add New Repository

```bash
# Add the new charts repository
helm repo add cronschedules https://cronschedules.github.io/charts
helm repo update
```

#### 3. Upgrade to New Chart

**Option A: In-place Upgrade (Recommended)**
```bash
# Upgrade using the new repository
helm upgrade cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --values my-values.yaml
```

**Option B: Fresh Installation**
```bash
# Uninstall old version
helm uninstall cronjob-scale-down-operator

# Install from new repository
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --values my-values.yaml
```

#### 4. Verify Migration

```bash
# Check installation
kubectl get pods -l app.kubernetes.io/name=cronjob-scale-down-operator

# Verify chart source
helm list -o json | jq '.[] | select(.name=="cronjob-scale-down-operator") | .chart'
```

## Benefits of Migration

### 🏗️ Infrastructure Improvements

- **Proper Hosting**: Charts hosted on GitHub Pages with automatic indexing
- **CI/CD Pipeline**: Automated testing, linting, and security scanning
- **Multiple Test Scenarios**: Charts tested with various configurations
- **Automated Releases**: Charts automatically packaged and released

### 📚 Documentation Enhancements

- **Dedicated Documentation**: Comprehensive chart-specific documentation
- **Configuration Examples**: Multiple values files for different use cases
- **Best Practices**: Following Helm chart repository standards
- **Migration Guides**: Clear migration paths and upgrade procedures

### 🔒 Security & Quality

- **Security Scanning**: Automated security scanning with Checkov
- **Lint Validation**: Comprehensive YAML and chart linting
- **Installation Testing**: Charts tested in real Kubernetes clusters
- **Version Management**: Proper semantic versioning and changelog

## Chart Documentation

For detailed chart documentation, visit the new charts repository:

- 📖 **Main Repository**: [cronschedules/charts](https://github.com/cronschedules/charts)
- 📋 **Chart Documentation**: [Chart README](https://github.com/cronschedules/charts/tree/main/cronjob-scale-down-operator)
- ⚙️ **Configuration Values**: [values.yaml](https://github.com/cronschedules/charts/blob/main/cronjob-scale-down-operator/values.yaml)
- 🧪 **Testing Examples**: [CI Values](https://github.com/cronschedules/charts/tree/main/cronjob-scale-down-operator/ci)

## Available Configurations

The new charts repository includes pre-configured values for different scenarios:

| Configuration | File | Description |
|---------------|------|-------------|
| **Default** | `values.yaml` | Standard production configuration |
| **Minimal** | `ci/minimal-values.yaml` | Resource-constrained environments |
| **High Availability** | `ci/ha-values.yaml` | Multi-replica production setup |
| **Testing** | `ci/testing-values.yaml` | CI/CD optimized configuration |

## Support

If you encounter any issues during migration:

1. **Check the Charts Repository**: [Issues](https://github.com/cronschedules/charts/issues)
2. **Migration Questions**: Create an issue in the [main repository](https://github.com/z4ck404/cronjob-scale-down-operator/issues)
3. **Chart Specific Issues**: Use the [charts repository](https://github.com/cronschedules/charts/issues)

## Frequently Asked Questions

### Q: Will the old chart continue to work?

A: The old chart location is deprecated and will not receive updates. We recommend migrating to the new repository for continued support and updates.

### Q: Can I still use local charts?

A: Yes, you can still clone the charts repository and use local charts:

```bash
git clone https://github.com/cronschedules/charts.git
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator
```

### Q: Are there any breaking changes?

A: The chart functionality remains the same. The main change is the repository location and hosting method.

### Q: What about CI/CD pipelines?

A: Update your CI/CD pipelines to use the new repository URL:

```yaml
# Before
helm install cronjob-operator ./charts/cronjob-scale-down-operator

# After  
helm repo add cronschedules https://cronschedules.github.io/charts
helm install cronjob-operator cronschedules/cronjob-scale-down-operator
```

---

**📅 Migration Timeline**: Charts were migrated in July 2025. The old location is now deprecated and will not receive updates.
