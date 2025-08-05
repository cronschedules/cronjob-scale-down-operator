# Helm Installation

Install the CronJob Scale Down Operator using Helm.

## Prerequisites

- Kubernetes cluster (v1.16+)
- Helm 3.0+
- kubectl configured

## Installation

```bash
# Add repository
helm repo add cronschedules https://cronschedules.github.io/charts
helm repo update

# Install
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator

# Or install with Docker Hub image
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --set image.repository=cronschedules/cronjob-scale-down-operator

# Verify
kubectl get pods -l app.kubernetes.io/name=cronjob-scale-down-operator
kubectl get crd cronjobscaledowns.cronschedules.elbazi.co
```

## Custom Configuration

### Override Values

Create `values.yaml`:

```yaml
image:
  tag: "0.4.0"  # Latest version with orphan cleanup
replicaCount: 2
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
webui:
  enabled: true
```

```bash
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator -f values.yaml
```

### Command Line Options

```bash
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --set image.tag=0.4.0 \
  --set replicaCount=2 \
  --set resources.requests.memory=128Mi
```

### Custom Namespace

```bash
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --namespace cronjob-operator \
  --create-namespace
```

## Upgrade

```bash
# Upgrade to latest
helm upgrade cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator

# Upgrade with values
helm upgrade cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator -f values.yaml
```

## Status

```bash
# Check release status
helm status cronjob-scale-down-operator

# List releases
helm list
```

## Uninstall

```bash
helm uninstall cronjob-scale-down-operator
```

## Local Development

For local chart development:

```bash
# Clone charts repository
git clone https://github.com/cronschedules/charts.git

# Lint chart
helm lint ./charts/cronjob-scale-down-operator

# Dry run
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator --dry-run

# Template rendering
helm template cronjob-scale-down-operator ./charts/cronjob-scale-down-operator
```

## Troubleshooting

**Installation fails:**
- Check Kubernetes version compatibility
- Verify RBAC permissions
- Check resource quotas

**Pods not starting:**
- Check image pull policy
- Verify resource requests/limits
- Check node capacity

```yaml
# Custom values for cronjob-scale-down-operator
image:
  tag: "0.4.0"

resources:
  limits:
    cpu: 1000m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

replicaCount: 2

nodeSelector:
  kubernetes.io/os: linux
```

Install with custom values:

```bash
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator -f values.yaml
```

### Install in Custom Namespace

```bash
# Create namespace
kubectl create namespace cronjob-operator

# Install in custom namespace
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --namespace cronjob-operator
```

### Install with Inline Values

```bash
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --set image.tag=0.4.0 \
  --set replicaCount=2 \
  --set resources.requests.memory=128Mi
```

## Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `ghcr.io/cronschedules/cronjob-scale-down-operator` |
| `image.tag` | Container image tag | `0.4.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `replicaCount` | Number of operator replicas | `1` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` |
| `rbac.create` | Create RBAC resources | `true` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `10m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `nodeSelector` | Node selector for pod assignment | `{}` |
| `tolerations` | Tolerations for pod assignment | `[]` |
| `affinity` | Affinity for pod assignment | `{}` |
| `metrics.enabled` | Enable metrics service | `true` |
| `leaderElection.enabled` | Enable leader election | `true` |

## Management Commands

### Upgrade

```bash
# Upgrade to latest version
helm upgrade cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator

# Upgrade with new values
helm upgrade cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator -f values.yaml
```

### Check Status

```bash
# Check Helm release status
helm status cronjob-scale-down-operator

# List all Helm releases
helm list
```

### Rollback

```bash
# See revision history
helm history cronjob-scale-down-operator

# Rollback to previous version
helm rollback cronjob-scale-down-operator 1
```

### Uninstall

```bash
# Uninstall the operator
helm uninstall cronjob-scale-down-operator

# Uninstall from custom namespace
helm uninstall cronjob-scale-down-operator --namespace cronjob-operator
```

## Troubleshooting

### Common Issues

1. **Helm not found:**
   ```bash
   # Install Helm
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
   ```

2. **Permission denied:**
   ```bash
   # Check if you have cluster admin permissions
   kubectl auth can-i create clusterroles
   ```

3. **Chart validation errors:**
   ```bash
   # For local development, clone the charts repository first:
   # git clone https://github.com/cronschedules/charts.git
   
   # Validate chart syntax
   helm lint ./charts/cronjob-scale-down-operator
   
   # Dry run installation
   helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator --dry-run
   ```

### Debug Commands

```bash
# Render templates locally (requires charts repository)
# git clone https://github.com/cronschedules/charts.git
helm template cronjob-scale-down-operator ./charts/cronjob-scale-down-operator

# Get rendered values
helm get values cronjob-scale-down-operator

# Get all resources created by Helm
helm get manifest cronjob-scale-down-operator
```

## Next Steps

After installation, you can:

1. **Create your first scaling schedule:**
   ```bash
   kubectl apply -f examples/quick-test.yaml
   ```

2. **Monitor the operator:**
   ```bash
   kubectl logs -l app.kubernetes.io/name=cronjob-scale-down-operator
   ```

3. **Check CronJobScaleDown resources:**
   ```bash
   kubectl get cronjobscaledown
   ```

For more examples and configuration options, see the [main documentation](../README.md).
