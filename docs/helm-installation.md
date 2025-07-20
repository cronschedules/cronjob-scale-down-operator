# Helm Installation Guide

This guide explains how to install the CronJob Scale Down Operator using Helm.

## Prerequisites

- Kubernetes cluster (v1.16+)
- Helm 3.0+
- kubectl configured with cluster admin permissions

## Installation

### Step 1: Clone the Repository

```bash
git clone https://github.com/z4ck404/cronjob-scale-down-operator.git
cd cronjob-scale-down-operator
```

### Step 2: Install the Chart

```bash
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator
```

### Step 3: Verify Installation

```bash
# Check if the operator is running
kubectl get pods -l app.kubernetes.io/name=cronjob-scale-down-operator

# Check if CRDs are installed
kubectl get crd cronjobscaledowns.cronschedules.elbazi.co
```

## Custom Configuration

### Override Default Values

Create a `values.yaml` file:

```yaml
# Custom values for cronjob-scale-down-operator
image:
  tag: "0.1.2"

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
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator -f values.yaml
```

### Install in Custom Namespace

```bash
# Create namespace
kubectl create namespace cronjob-operator

# Install in custom namespace
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator \
  --namespace cronjob-operator
```

### Install with Inline Values

```bash
helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator \
  --set image.tag=0.1.2 \
  --set replicaCount=2 \
  --set resources.requests.memory=128Mi
```

## Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `ghcr.io/z4ck404/cronjob-scale-down-operator` |
| `image.tag` | Container image tag | `0.1.2` |
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
helm upgrade cronjob-scale-down-operator ./charts/cronjob-scale-down-operator

# Upgrade with new values
helm upgrade cronjob-scale-down-operator ./charts/cronjob-scale-down-operator -f values.yaml
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
   # Validate chart syntax
   helm lint ./charts/cronjob-scale-down-operator
   
   # Dry run installation
   helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator --dry-run
   ```

### Debug Commands

```bash
# Render templates locally
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
