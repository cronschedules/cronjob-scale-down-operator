# CronJob Scale Down Operator

Kubernetes operator for scheduled scaling of Deployments and StatefulSets.

## Features

- Cron-based scheduling with second precision
- Timezone support
- Scale down/up on different schedules
- Supports Deployments and StatefulSets
- Resource cleanup based on annotations
- Cleanup-only mode (no scaling target)
- Status tracking
- Web UI dashboard
- Dry-run mode for testing

## Installation

### Helm (Recommended)

```bash
helm repo add cronschedules https://cronschedules.github.io/charts
helm repo update
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator
```

### Manual

```bash
kubectl apply -f https://github.com/z4ck404/cronjob-scale-down-operator/releases/latest/download/install.yaml
```

## Quick Test

1. **Create test deployment:**
   ```bash
   kubectl apply -f examples/test-deployment.yaml
   ```

2. **Apply scaling schedule:**
   ```bash
   kubectl apply -f examples/quick-test.yaml
   ```

3. **Monitor:**
   ```bash
   kubectl get cronjobscaledown -w
   ```

## Basic Usage

### Simple Scaling

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: nightly-scaling
spec:
  targetRef:
    name: my-deployment
    kind: Deployment
    apiVersion: apps/v1
  scaleDownSchedule: "0 0 22 * * *"  # 10 PM
  scaleUpSchedule: "0 0 8 * * *"     # 8 AM
  timeZone: "America/New_York"
```

### Cleanup Only

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: cleanup-job
spec:
  cleanupSchedule: "0 0 2 * * *"  # 2 AM daily
  cleanupConfig:
    annotationKey: "test.example.com/cleanup-after"
    resourceTypes: ["ConfigMap", "Secret", "Pod"]
```

### Combined Scaling + Cleanup

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: full-management
spec:
  targetRef:
    name: my-app
    kind: Deployment
    apiVersion: apps/v1
  scaleDownSchedule: "0 0 18 * * *"
  scaleUpSchedule: "0 0 9 * * *"
  cleanupSchedule: "0 0 1 * * *"
  cleanupConfig:
    annotationKey: "test.example.com/cleanup-after"
    resourceTypes: ["ConfigMap", "Secret"]
  timeZone: "UTC"
```

## Configuration

### Schedule Format

Uses standard cron with seconds:
```
# ┌───────────── second (0 - 59)
# │ ┌───────────── minute (0 - 59)
# │ │ ┌───────────── hour (0 - 23)
# │ │ │ ┌───────────── day of month (1 - 31)
# │ │ │ │ ┌───────────── month (1 - 12)
# │ │ │ │ │ ┌───────────── day of week (0 - 6)
# │ │ │ │ │ │
# * * * * * *
```

Examples:
- `0 0 22 * * *` - Daily at 10 PM
- `0 0 18 * * 1-5` - Weekdays at 6 PM
- `0 0 9 * * 1` - Mondays at 9 AM

### Timezones

Supports all IANA timezone names:
- `UTC`
- `America/New_York`
- `Europe/London`
- `Asia/Tokyo`

## Status Monitoring

```bash
# Check status
kubectl get cronjobscaledown

# Detailed view
kubectl describe cronjobscaledown my-job

# Watch for changes
kubectl get cronjobscaledown -w
```

## Web UI

Access the dashboard at http://localhost:8082 (when running locally) or configure ingress for production.

```bash
# Enable in Helm
helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
  --set webui.enabled=true
```

## Documentation

- [Helm Installation](helm-installation.md)
- [Web UI Guide](webui.md)
- [Cleanup Feature](cleanup.md)
- [Charts Migration](charts-migration.md)

> **📖 Chart Documentation:** For detailed Helm chart documentation, values, and configuration options, visit the [Charts Repository](https://github.com/cronschedules/charts/tree/main/cronjob-scale-down-operator).

#### Option 2: Using Container Image

The operator is available as a pre-built container image from multiple registries:

```bash
# From Docker Hub:
docker pull cronschedules/cronjob-scale-down-operator:0.3.0

# From GitHub Container Registry:
docker pull ghcr.io/cronschedules/cronjob-scale-down-operator:0.3.0
```

Use these images in your custom deployments or with the provided Helm chart.

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

The [`examples/`](../examples/) directory contains various use cases:

| Example | Description | Schedule |
|---------|-------------|----------|
| **[quick-test.yaml](../examples/quick-test.yaml)** | Immediate testing | Every minute |
| **[basic-daily-schedule.yaml](../examples/basic-daily-schedule.yaml)** | Production workload | 10 PM → 6 AM daily |
| **[weekend-shutdown.yaml](../examples/weekend-shutdown.yaml)** | Weekend cost savings | Friday 6 PM → Monday 8 AM |
| **[development-testing.yaml](../examples/development-testing.yaml)** | Dev environment | Every 30/45 seconds |
| **[multi-timezone.yaml](../examples/multi-timezone.yaml)** | Global deployments | Multiple timezones |
| **[statefulset-example.yaml](../examples/statefulset-example.yaml)** | Database scaling | StatefulSet support |
| **[cleanup-only-example.yaml](../examples/cleanup-only-example.yaml)** | **Cleanup only** | **Every 6 hours** |

## Cleanup-Only Mode

The operator supports a cleanup-only mode where it manages resource cleanup without scaling any target resources. This is perfect for environments where you need automated cleanup of test resources, temporary objects, or expired configurations.

### When to Use Cleanup-Only Mode

- **CI/CD Pipelines**: Automatically clean up test resources after builds
- **Development Environments**: Remove temporary test objects on a schedule
- **Resource Management**: Clean up expired ConfigMaps, Secrets, or test deployments
- **Cost Optimization**: Remove unused resources to save cluster costs

### Cleanup-Only Configuration

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
    namespaces:
      - "test"
      - "staging"
    labelSelector:
      environment: "test"
    dryRun: false
  timeZone: "UTC"
```

### Combined Scaling + Cleanup

You can also combine scaling and cleanup in a single resource:

```yaml
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: combined-scaler-cleanup
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
    resourceTypes: ["ConfigMap", "Secret"]
    dryRun: false
  timeZone: "UTC"
```

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

## Helm Chart Installation

The operator can be installed using Helm for easier management and configuration:

### Chart Information

- **Repository**: `ghcr.io/cronschedules/cronjob-scale-down-operator` or `cronschedules/cronjob-scale-down-operator` (Docker Hub)
- **Image Tag**: `0.3.0`
- **Chart Version**: `0.3.0`

### Installation Steps

1. **Add repository and install:**
   ```bash
   helm repo add cronschedules https://cronschedules.github.io/charts
   helm repo update
   helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator
   ```

2. **Custom values:**
   ```bash
   helm install cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator \
     --set image.tag=0.3.0 \
     --set resources.requests.memory=128Mi \
     --set replicaCount=1
   ```

3. **Upgrade:**
   ```bash
   helm upgrade cronjob-scale-down-operator cronschedules/cronjob-scale-down-operator
   ```

4. **Uninstall:**
   ```bash
   helm uninstall cronjob-scale-down-operator
   ```

### Chart Configuration

Key Helm chart values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `ghcr.io/cronschedules/cronjob-scale-down-operator` |
| `image.tag` | Container image tag | `0.3.0` |
| `replicaCount` | Number of operator replicas | `1` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `10m` |
| `metrics.enabled` | Enable metrics service | `true` |
| `rbac.create` | Create RBAC resources | `true` |

## Web UI Dashboard

The operator includes a built-in web dashboard that provides real-time monitoring of all CronJobScaleDown resources and their target deployments/statefulsets.

![Web UI Dashboard](./images/web-ui.png)

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

For more details about the web UI, see the [Web UI Documentation](./webui.md).

## Monitoring

### Check CronJobScaleDown Status

```bash
kubectl get cronjobscaledown -o wide
kubectl describe cronjobscaledown my-scaler
```

### View Operator Logs

```bash
kubectl logs -l app.kubernetes.io/name=cronjob-scale-down-operator
```

### Monitor Target Resources

```bash
kubectl get deployment my-deployment -w
kubectl get statefulset my-statefulset -w
```

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
kubectl get deployment -l app.kubernetes.io/name=cronjob-scale-down-operator

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

Licensed under the Apache License, Version 2.0. See [LICENSE](../LICENSE) for details.

## Support

- 📧 **Issues**: [GitHub Issues](https://github.com/z4ck404/cronjob-scale-down-operator/issues)
- 📖 **Documentation**: [docs/](.)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/z4ck404/cronjob-scale-down-operator/discussions)