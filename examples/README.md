# CronJobScaleDown Examples

This directory contains various examples demonstrating different use cases for the CronJobScaleDown operator.

## Examples Overview

| Example | Description | Use Case |
|---------|-------------|----------|
| `basic-daily-schedule.yaml` | Basic daily scaling schedule | Production workloads with night/weekend downtime |
| `development-testing.yaml` | Frequent scaling for testing | Development and testing environments |
| `quick-test.yaml` | Immediate scaling test | Quick validation and testing |
| `weekend-shutdown.yaml` | Weekend-only scaling | Cost optimization for non-critical services |
| `multi-timezone.yaml` | Different timezone examples | Global deployments |
| `statefulset-example.yaml` | StatefulSet scaling example | Database and stateful application scaling |
| `cleanup-only-example.yaml` | **Cleanup-only mode** | **Pure resource cleanup without scaling** |
| `webui-demo.yaml` | Web UI demonstration | Complete example with deployment and scaling |

## Cleanup Examples

The operator now supports automated resource cleanup based on annotations:

- **`cleanup-only-example.yaml`**: Demonstrates cleanup-only mode where the operator only performs cleanup operations without scaling any target resources
- **Combined examples**: Most scaling examples can be extended with cleanup configuration

### Test Resource Cleanup

To test the cleanup functionality:

1. **Apply test resources with cleanup annotations:**
   ```bash
   kubectl apply -f ../test-cleanup-resources.yaml
   ```

2. **Apply a cleanup-only job:**
   ```bash
   kubectl apply -f cleanup-only-example.yaml
   ```

3. **Monitor cleanup operations:**
   ```bash
   kubectl logs -n cronjob-scale-down-operator-system deployment/cronjob-scale-down-operator-controller-manager | grep -i cleanup
   ```

## Testing Your Setup

1. **Create a test deployment:**
   ```bash
   kubectl create deployment nginx-test --image=nginx --replicas=3
   ```

2. **Apply one of the examples:**
   ```bash
   kubectl apply -f examples/quick-test.yaml
   ```

3. **Monitor the scaling:**
   ```bash
   kubectl get cronjobscaledown -w
   kubectl get deployment nginx-test -w
   ```

## Schedule Format

The operator supports cron expressions with seconds precision:
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

### Common Schedule Examples

- `"0 0 22 * * *"` - Every day at 10:00 PM
- `"0 0 6 * * *"` - Every day at 6:00 AM  
- `"0 0 0 * * 0"` - Every Sunday at midnight
- `"*/30 * * * * *"` - Every 30 seconds (testing)
- `"0 0 18 * * 1-5"` - Weekdays at 6:00 PM
- `"0 0 8 * * 1-5"` - Weekdays at 8:00 AM
