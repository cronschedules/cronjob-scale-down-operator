# Resource Cleanup

Automatically clean up Kubernetes resources based on annotations and schedules.

## Overview

Two modes:
- **Cleanup-Only**: Pure cleanup without scaling
- **Combined**: Cleanup + scaling operations

## Cleanup-Only Mode

For automated cleanup without target scaling.

### Use Cases

- CI/CD pipeline cleanup
- Development environment cleanup
- Test resource management
- Cost optimization

### Configuration

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: cleanup-only-job
spec:
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

## Resource Types

Supported resource types for cleanup:

```yaml
cleanupConfig:
  resourceTypes:
    - "ConfigMap"
    - "Secret"
    - "Service"
    - "Deployment"
    - "StatefulSet"
    - "Job"
    - "Pod"
```

## Annotation-Based Cleanup

Resources with specific annotations are cleaned up when their timestamp expires.

### Format

```yaml
annotations:
  test.example.com/cleanup-after: "2025-01-20T15:30:00Z"
```

### Example Resource

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  annotations:
    test.example.com/cleanup-after: "2025-01-20T15:30:00Z"
data:
  test: "data"
```

## Combined Mode

Cleanup alongside scaling operations:

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: combined-job
spec:
  targetRef:
    name: my-deployment
    kind: Deployment
    apiVersion: apps/v1
  scaleDownSchedule: "0 0 22 * * *"
  scaleUpSchedule: "0 0 8 * * *" 
  cleanupSchedule: "0 0 2 * * *"
  cleanupConfig:
    annotationKey: "test.example.com/cleanup-after"
    resourceTypes: ["ConfigMap", "Secret"]
```

## Dry Run Mode

Test cleanup operations without deleting resources:

```yaml
cleanupConfig:
  dryRun: true
```

Dry run logs what would be deleted without performing actual deletions.

## Examples

### CI/CD Cleanup

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: ci-cleanup
spec:
  cleanupSchedule: "0 */30 * * * *"  # Every 30 minutes
  cleanupConfig:
    annotationKey: "ci.example.com/cleanup-after"
    resourceTypes: ["Pod", "ConfigMap", "Secret"]
    dryRun: false
```

### Development Environment

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: dev-cleanup
spec:
  cleanupSchedule: "0 0 1 * * *"  # Daily at 1 AM
  cleanupConfig:
    annotationKey: "dev.example.com/cleanup-after"
    resourceTypes: ["Deployment", "Service", "ConfigMap"]
```

## Status

Check cleanup status:

```bash
kubectl get cronjobscaledown cleanup-only-job -o yaml
```

View last cleanup execution in status field.
