![Logo](./docs/images/logo.png)
# CronJob-Scale-Down-Operator

A Kubernetes operator that automatically scales down Deployments and StatefulSets during specific time windows (e.g., at night or on weekends) to save resources and costs.

## Features

- 🕒 **Cron-based Scheduling**: Uses standard cron expressions with second precision
- 🌍 **Timezone Support**: Configure schedules in any timezone
- 📈 **Flexible Scaling**: Scale down and up on different schedules
- 🎯 **Multiple Resource Types**: Supports Deployments and StatefulSets for scaling
- 🧹 **Resource Cleanup**: Automatically delete test resources based on annotations
- 📊 **Status Tracking**: Monitor last execution times and current replica counts
- 🌐 **Web UI Dashboard**: Built-in web interface to monitor all cron jobs and their status
- ⚡ **Efficient**: Only reconciles when needed, with smart requeue timing
- 🛡️ **Safe Testing**: Dry-run mode for cleanup operations

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.16+)
- kubectl configured
- Cluster admin permissions

### Installation

#### Option 1: Using Helm (Recommended)

1. **Install using Helm:**
   ```bash
   # Clone the repository
   git clone https://github.com/z4ck404/cronjob-scale-down-operator.git
   cd cronjob-scale-down-operator
   
   # Install the operator
   helm install cronjob-scale-down-operator ./charts/cronjob-scale-down-operator
   ```

2. **Verify installation:**
   ```bash
   kubectl get pods -l app.kubernetes.io/name=cronjob-scale-down-operator
   ```

#### Option 2: Using Container Image

The operator is available as a pre-built container image:

```bash
# Image available at:
docker pull ghcr.io/z4ck404/cronjob-scale-down-operator:0.1.2
```

Use this image in your custom deployments or with the provided Helm chart.

#### Option 3: Using kubectl

1. **Install the CRDs and operator:**
   ```bash
   kubectl apply -f config/crd/bases/
   kubectl apply -f config/rbac/
   kubectl apply -f config/manager/
   ```

#### Quick Test

1. **Create a test deployment:**
   ```bash
   kubectl apply -f examples/test-deployment.yaml
   ```

2. **Apply a scaling schedule:**
   ```bash
   kubectl apply -f examples/quick-test.yaml
   ```

3. **Monitor the scaling:**
   ```bash
   kubectl get cronjobscaledown -w
   kubectl get deployment nginx-test -w
   ```

## Examples

The [`examples/`](./examples/) directory contains various use cases:

| Example | Description | Schedule |
|---------|-------------|----------|
| **[quick-test.yaml](./examples/quick-test.yaml)** | Immediate testing | Every minute |
| **[basic-daily-schedule.yaml](./examples/basic-daily-schedule.yaml)** | Production workload | 10 PM → 6 AM daily |
| **[weekend-shutdown.yaml](./examples/weekend-shutdown.yaml)** | Weekend cost savings | Friday 6 PM → Monday 8 AM |
| **[development-testing.yaml](./examples/development-testing.yaml)** | Dev environment | Every 30/45 seconds |
| **[multi-timezone.yaml](./examples/multi-timezone.yaml)** | Global deployments | Multiple timezones |
| **[statefulset-example.yaml](./examples/statefulset-example.yaml)** | Database scaling | StatefulSet support |
| **[resource-cleanup-example.yaml](./examples/resource-cleanup-example.yaml)** | **Resource cleanup** | **Combined scaling + cleanup** |
| **[cleanup-only-example.yaml](./examples/cleanup-only-example.yaml)** | **Cleanup only** | **Every 6 hours** |

## Configuration

### CronJobScaleDown Spec

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: my-scaler
  namespace: default
spec:
  # Target resource to scale
  targetRef:
    name: my-deployment
    namespace: default
    kind: Deployment  # or StatefulSet
    apiVersion: apps/v1
  
  # When to scale down (cron format with seconds)
  scaleDownSchedule: "0 0 22 * * *"  # 10 PM daily
  
  # When to scale up (optional)
  scaleUpSchedule: "0 0 6 * * *"     # 6 AM daily
  
  # Timezone for schedule interpretation
  timeZone: "UTC"  # or "America/New_York", "Europe/London", etc.
```

### Resource Cleanup Configuration

The operator now supports automatic cleanup of test resources based on annotations:

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: test-cleanup
  namespace: default
spec:
  # Optional: You can still use scaling features alongside cleanup
  targetRef:
    name: my-deployment
    namespace: default
    kind: Deployment
    apiVersion: apps/v1
  scaleDownSchedule: "0 0 22 * * *"
  scaleUpSchedule: "0 0 6 * * *"
  
  # Cleanup configuration
  cleanupSchedule: "0 0 2 * * *"  # Run cleanup at 2 AM daily
  cleanupConfig:
    # Annotation key that marks resources for cleanup
    annotationKey: "cleanup-after"
    
    # Resource types to cleanup
    resourceTypes:
      - "Deployment"
      - "Service" 
      - "ConfigMap"
      - "Secret"
      - "StatefulSet"
    
    # Optional: Limit cleanup to specific namespaces
    namespaces:
      - "test"
      - "lab"
    
    # Optional: Additional label selector
    labelSelector:
      environment: "test"
    
    # Optional: Enable dry-run mode (default: false)
    dryRun: false
  
  timeZone: "UTC"
```

#### Cleanup Time Formats

Resources can be marked for cleanup using various time formats in the annotation value:

- **Duration**: `24h`, `7d`, `30m` (relative to resource creation time)
- **Absolute time**: `2024-12-31T23:59:59Z` (RFC3339 format)
- **Date**: `2024-12-31` (cleanup at midnight on that date)
- **Immediate**: Empty value `""` (cleanup on next schedule run)

#### Example Resource with Cleanup Annotation

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test
  labels:
    environment: "test"
  annotations:
    cleanup-after: "24h"  # Delete 24 hours after creation
spec:
  # ... deployment spec
```

### Schedule Format

The operator supports 6-field cron expressions with second precision:

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

#### Common Schedule Examples

| Schedule | Description |
|----------|-------------|
| `"0 0 22 * * *"` | Every day at 10:00 PM |
| `"0 0 6 * * 1-5"` | Weekdays at 6:00 AM |
| `"0 0 18 * * 5"` | Every Friday at 6:00 PM |
| `"0 0 0 * * 0"` | Every Sunday at midnight |
| `"*/30 * * * * *"` | Every 30 seconds (testing) |

### Supported Timezones

Use standard IANA timezone names:
- `UTC`
- `America/New_York`
- `Europe/London`
- `Europe/Berlin`
- `Asia/Tokyo`
- `Australia/Sydney`

## Monitoring

### Check CronJobScaleDown Status

```bash
kubectl get cronjobscaledown -o wide
kubectl describe cronjobscaledown my-scaler
```

### View Operator Logs

```bash
kubectl logs -n cronjob-scale-down-operator-system deployment/cronjob-scale-down-operator-controller-manager
```

### Monitor Target Resources

```bash
kubectl get deployment my-deployment -w
kubectl get statefulset my-statefulset -w
```

## Web UI Dashboard

The operator includes a built-in web dashboard that provides real-time monitoring of all CronJobScaleDown resources and their target deployments/statefulsets.

![Web UI Dashboard](./docs/images/web-ui.png)

### Accessing the Web UI

By default, the web UI is available at `http://localhost:8082` when running the operator locally. In a Kubernetes cluster, you can access it by:

1. **Port forwarding** (for development/testing):
   ```bash
   kubectl port-forward -n cronjob-scale-down-operator-system deployment/cronjob-scale-down-operator-controller-manager 8082:8082
   ```
   Then visit `http://localhost:8082`

2. **Configure ingress** (for production):
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: cronjob-scale-down-operator-ui
   spec:
     rules:
     - host: cronjob-ui.example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: cronjob-scale-down-operator-ui
               port:
                 number: 8082
   ```

### Web UI Features

- 📊 **Real-time Dashboard**: Overview of all CronJobScaleDown resources
- 📈 **Status Monitoring**: Current state of target deployments and statefulsets  
- 🕒 **Schedule Information**: View scale-up/down schedules and timezones
- 📋 **Replica Status**: Visual indicators for ready vs desired replicas
- 📅 **Action History**: Timestamps of last scale operations
- 🔄 **Auto-refresh**: Updates every 30 seconds automatically
- 📱 **Responsive Design**: Works on desktop, tablet, and mobile

### Customizing Web UI Port

You can customize the web UI port using the `--webui-addr` flag:

```bash
./manager --webui-addr=:8080
```

For more details about the web UI, see the [Web UI Documentation](./docs/webui.md).

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/z4ck404/cronjob-scale-down-operator.git
cd cronjob-scale-down-operator

# Build and run locally
make run

# Build Docker image
make docker-build IMG=my-registry/cronjob-scale-down-operator:latest

# Deploy to cluster
make deploy IMG=my-registry/cronjob-scale-down-operator:latest
```

### Running Tests

```bash
# Unit tests
make test

# End-to-end tests
make test-e2e
```

## Troubleshooting

### Common Issues

1. **Scaling not happening:**
   - Check timezone configuration
   - Verify cron schedule syntax
   - Check operator logs for errors

2. **Permission errors:**
   - Ensure RBAC is properly configured
   - Verify service account permissions

3. **Target resource not found:**
   - Check namespace and resource name
   - Verify resource exists and is accessible

### Debug Commands

```bash
# Check CRD installation
kubectl get crd cronjobscaledowns.cronschedules.elbazi.co

# Verify operator deployment
kubectl get deployment -n cronjob-scale-down-operator-system

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Support

- 📧 **Issues**: [GitHub Issues](https://github.com/z4ck404/cronjob-scale-down-operator/issues)
- 📖 **Documentation**: [docs/](./docs/)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/z4ck404/cronjob-scale-down-operator/discussions)