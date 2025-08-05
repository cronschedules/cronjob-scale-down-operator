# Resource Cleanup

Automatically clean up Kubernetes resources based on annotations, schedules, and age thresholds.

## Overview

The CronJob Scale Down Operator provides comprehensive resource cleanup capabilities with two main approaches:

- **Annotation-Based Cleanup**: Resources with specific cleanup annotations
- **Orphan Resource Cleanup**: Resources without annotations based on age thresholds (v0.4.0+)

### Deployment Modes

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
    
    # NEW in v0.4.0: Orphan resource cleanup
    cleanupOrphanResources: true
    orphanResourceMaxAge: "168h"  # 7 days
    
    # Optional: Target specific resources
    labelSelector:
      app.kubernetes.io/managed-by: "test"
      
    dryRun: false
  timeZone: "UTC"
```

## Resource Types

Supported resource types for cleanup (expanded in v0.4.0):

```yaml
cleanupConfig:
  resourceTypes:
    # Standard Resources
    - "ConfigMap"
    - "Secret"
    - "Service"
    - "Deployment"
    - "StatefulSet"
    
    # Workload Resources (v0.4.0+)
    - "Job"
    - "Pod"
    
    # RBAC Resources (v0.4.0+)
    - "Role"
    - "RoleBinding"
    - "ClusterRole"
    - "ClusterRoleBinding"
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

## Orphan Resource Cleanup (v0.4.0+)

Clean up resources **without** cleanup annotations based on age thresholds. Perfect for cleaning up forgotten test resources and CI/CD artifacts.

### Configuration

```yaml
cleanupConfig:
  # Enable orphan cleanup (disabled by default)
  cleanupOrphanResources: true
  
  # Age threshold - resources older than this will be cleaned
  orphanResourceMaxAge: "72h"  # 3 days
  
  # Optional: Only clean resources matching these labels
  labelSelector:
    app.kubernetes.io/created-by: "test"
    environment: "development"
```

### How It Works

1. **Standard Cleanup**: Resources WITH annotations are cleaned based on their timestamp
2. **Orphan Cleanup**: Resources WITHOUT annotations but matching criteria are cleaned based on age
3. **Safety First**: Resources must be older than `orphanResourceMaxAge` threshold
4. **Label Filtering**: Only resources matching `labelSelector` are considered for orphan cleanup

### Use Cases

- **CI/CD Cleanup**: Remove test artifacts older than a specific age
- **Development Environment**: Clean up forgotten development resources  
- **Failed Job Cleanup**: Remove failed pods and jobs automatically
- **RBAC Cleanup**: Clean up temporary roles and bindings

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
    resourceTypes: ["Pod", "ConfigMap", "Secret", "Job"]
    
    # Clean orphan CI resources older than 2 hours
    cleanupOrphanResources: true
    orphanResourceMaxAge: "2h"
    
    labelSelector:
      ci-pipeline: "true"
    
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
    
    # Clean forgotten dev resources older than 3 days
    cleanupOrphanResources: true
    orphanResourceMaxAge: "72h"
    
    labelSelector:
      environment: "development"
```

### Failed Workload Cleanup

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: failed-workload-cleanup
spec:
  cleanupSchedule: "0 0 */4 * * *"  # Every 4 hours
  cleanupConfig:
    resourceTypes: ["Pod", "Job"]
    
    # Only orphan cleanup for failed workloads
    cleanupOrphanResources: true
    orphanResourceMaxAge: "1h"
    
    labelSelector:
      job-type: "batch"
```

## Status

Check cleanup status:

```bash
kubectl get cronjobscaledown cleanup-only-job -o yaml
```

View last cleanup execution in status field.
