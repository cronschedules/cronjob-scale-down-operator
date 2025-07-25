# Charts Migration

Helm charts moved to dedicated repository: `cronschedules/charts`

## Changes

| Detail | Old | New |
|--------|-----|-----|
| **Repository** | `z4ck404/cronjob-scale-down-operator` | `cronschedules/charts` |
| **Location** | `/charts/cronjob-scale-down-operator/` | `/cronjob-scale-down-operator/` |
| **Helm Repo** | Local only | `https://cronschedules.github.io/charts` |

## New Installation

```bash
helm repo add cronschedules https://cronschedules.github.io/charts
helm repo update
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator
```

## Migration Steps

### 1. Backup Current Installation

```bash
helm get values cronjob-scale-down-operator > my-values.yaml
```

### 2. Uninstall Old Chart

```bash
helm uninstall cronjob-scale-down-operator
```

### 3. Install New Chart

```bash
helm repo add cronschedules https://cronschedules.github.io/charts
helm repo update
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator -f my-values.yaml
```

## Local Development

For local chart development, clone the charts repository:

```bash
git clone https://github.com/cronschedules/charts.git
helm install my-operator ./charts/cronjob-scale-down-operator
```

## FAQ

**Q: Why migrate charts?**
A: Better hosting, automated CI/CD, security scanning, and proper Helm repository practices.

**Q: Are there breaking changes?**
A: No functional changes. Only repository location changed.

**Q: Can I use local charts?**
A: Yes, clone the charts repository and install locally.

**Q: What about CI/CD pipelines?**
A: Update to use the new repository URL:

```yaml
# Before
helm install cronjob-operator ./charts/cronjob-scale-down-operator

# After  
helm repo add cronschedules https://cronschedules.github.io/charts
helm install cronjob-operator cronschedules/cronjob-scale-down-operator
```

## Support

- [Charts Repository Issues](https://github.com/cronschedules/charts/issues)
- [Operator Issues](https://github.com/z4ck404/cronjob-scale-down-operator/issues)
