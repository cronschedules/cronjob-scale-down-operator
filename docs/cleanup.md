# Resource Cleanup Feature

The CronJob Scale Down Operator provides comprehensive resource cleanup capabilities that can be used standalone or in combination with scaling features. This allows you to automatically clean up test resources, expired configurations, and temporary objects based on annotations and schedules.

## Overview

The cleanup feature supports two modes:
- **Cleanup-Only Mode**: Pure cleanup functionality without any scaling target
- **Combined Mode**: Cleanup alongside existing scaling operations

## Cleanup-Only Mode

Cleanup-only mode is perfect for environments where you need automated cleanup without scaling any target resources.

### Use Cases

- **CI/CD Pipelines**: Automatically clean up test resources after builds complete
- **Development Environments**: Remove temporary test objects on a schedule
- **Resource Management**: Clean up expired ConfigMaps, Secrets, or test deployments
- **Cost Optimization**: Remove unused resources to save cluster costs
- **Compliance**: Ensure test data is removed according to data retention policies

### Basic Configuration

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: cleanup-only-job
  namespace: default
spec:
  # No targetRef needed for cleanup-only mode
  cleanupSchedule: "0 0 */6 * * *"  # Every 6 hours
  cleanupConfig:
    annotationKey: "test.example.com/cleanup-after"
    resourceTypes:
      - "ConfigMap"
      - "Secret"
      - "Service"
      - "Deployment"
    dryRun: false
  timeZone: "UTC"
```

## Cleanup Configuration Options

### Resource Types

Specify which Kubernetes resource types to include in cleanup operations:

```yaml
cleanupConfig:
  resourceTypes:
    - "Deployment"
    - "Service"
    - "ConfigMap"
    - "Secret"
    - "StatefulSet"
    - "Job"
    - "CronJob"
    - "Ingress"
    - "PersistentVolumeClaim"
```

### Namespace Filtering

Limit cleanup operations to specific namespaces:

```yaml
cleanupConfig:
  namespaces:
    - "test"
    - "staging"
    - "development"
    # If empty, searches all accessible namespaces
```

### Label Selectors

Use label selectors to target specific resources:

```yaml
cleanupConfig:
  labelSelector:
    environment: "test"
    project: "my-project"
    temporary: "true"
```

### Annotation-Based Cleanup

Resources are marked for cleanup using annotations with various time formats:

#### Time Format Examples

```yaml
# Duration relative to resource creation time
cleanup-after: "24h"    # 24 hours after creation
cleanup-after: "7d"     # 7 days after creation
cleanup-after: "30m"    # 30 minutes after creation

# Absolute timestamp (RFC3339 format)
cleanup-after: "2024-12-31T23:59:59Z"

# Date (cleanup at midnight UTC)
cleanup-after: "2024-12-31"

# Immediate cleanup (on next schedule run)
cleanup-after: ""
```

### Example Resource with Cleanup Annotation

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: test
  labels:
    environment: "test"
    project: "my-app"
  annotations:
    test.example.com/cleanup-after: "24h"
data:
  config.json: |
    {"test": "data"}
```

## Combined Scaling + Cleanup

You can combine scaling and cleanup in a single CronJobScaleDown resource:

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: app-scaler-with-cleanup
  namespace: default
spec:
  # Scaling configuration
  targetRef:
    name: my-app
    namespace: default
    kind: Deployment
    apiVersion: apps/v1
  scaleDownSchedule: "0 0 22 * * *"  # Scale down at 10 PM
  scaleUpSchedule: "0 0 6 * * *"     # Scale up at 6 AM
  
  # Cleanup configuration
  cleanupSchedule: "0 0 2 * * *"     # Clean up at 2 AM
  cleanupConfig:
    annotationKey: "cleanup-after"
    resourceTypes:
      - "ConfigMap"
      - "Secret"
    labelSelector:
      app: "my-app"
    dryRun: false
  timeZone: "UTC"
```

## Dry Run Mode

Enable dry-run mode to see what would be deleted without actually deleting resources:

```yaml
cleanupConfig:
  dryRun: true  # Only log what would be deleted
```

When dry-run is enabled, the operator will:
- Log all resources that match the cleanup criteria
- Show what would be deleted in the operator logs
- Not actually delete any resources
- Update the cleanup timestamp as if cleanup occurred

## Schedule Format

Cleanup schedules use the same 6-field cron format as scaling schedules:

```
┌─────────────second (0 - 59)
│ ┌───────────── minute (0 - 59)
│ │ ┌───────────── hour (0 - 23)
│ │ │ ┌───────────── day of month (1 - 31)
│ │ │ │ ┌───────────── month (1 - 12)
│ │ │ │ │ ┌───────────── day of week (0 - 6) (0 = Sunday)
│ │ │ │ │ │
* * * * * *
```

### Common Cleanup Schedules

| Schedule | Description |
|----------|-------------|
| `"0 0 2 * * *"` | Every day at 2:00 AM |
| `"0 0 */6 * * *"` | Every 6 hours |
| `"0 0 0 * * 0"` | Every Sunday at midnight |
| `"0 0 22 * * 5"` | Every Friday at 10:00 PM |
| `"0 */30 * * * *"` | Every 30 minutes |

## Monitoring and Observability

### Check Cleanup Status

```bash
# View cleanup-only resources
kubectl get cronjobscaledown -o jsonpath='{.items[?(@.spec.cleanupSchedule)].metadata.name}'

# Check last cleanup time
kubectl get cronjobscaledown my-cleanup -o jsonpath='{.status.lastCleanupTime}'

# View detailed status
kubectl describe cronjobscaledown my-cleanup
```

### Operator Logs

Monitor cleanup operations in the operator logs:

```bash
kubectl logs -n cronjob-scale-down-operator-system \
  deployment/cronjob-scale-down-operator-controller-manager \
  | grep -i cleanup
```

### Web UI Dashboard

The built-in web UI provides special views for cleanup-only resources:
- Cleanup-only badge for easy identification
- Cleanup schedule display
- Last cleanup timestamp
- No target resource information (since there isn't one)

## Security Considerations

### RBAC Permissions

Ensure the operator has appropriate permissions for cleanup operations:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cronjob-scale-down-operator-cleanup
rules:
# Add list, get, delete permissions for each resource type
- apiGroups: [""]
  resources: ["configmaps", "secrets", "services"]
  verbs: ["list", "get", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets"]
  verbs: ["list", "get", "delete"]
```

### Cleanup Safety

- **Test with dry-run**: Always test cleanup configurations with `dryRun: true` first
- **Use specific selectors**: Avoid overly broad label selectors or namespace selections
- **Monitor logs**: Watch operator logs during initial cleanup runs
- **Backup important data**: Ensure critical resources are not accidentally marked for cleanup

## Troubleshooting

### Common Issues

1. **Resources not being cleaned up**
   - Check annotation format and values
   - Verify resource matches labelSelector and namespace filters
   - Ensure cleanup time has passed
   - Check operator logs for errors

2. **Permission denied errors**
   - Verify RBAC permissions for target resource types
   - Check if operator can access target namespaces

3. **Dry run not working**
   - Ensure `dryRun: true` is set in cleanupConfig
   - Check operator logs for dry-run messages

### Debug Commands

```bash
# Check cleanup configuration
kubectl get cronjobscaledown my-cleanup -o yaml

# View operator logs
kubectl logs -n cronjob-scale-down-operator-system \
  deployment/cronjob-scale-down-operator-controller-manager

# List resources with cleanup annotations
kubectl get configmaps --all-namespaces \
  -o jsonpath='{.items[?(@.metadata.annotations.cleanup-after)].metadata.name}'
```

## Examples

See the [`examples/`](../examples/) directory for complete working examples:

- [`cleanup-only-example.yaml`](../examples/cleanup-only-example.yaml) - Pure cleanup-only configuration
- [`test-cleanup-resources.yaml`](../test-cleanup-resources.yaml) - Test resources with cleanup annotations
