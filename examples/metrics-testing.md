# Metrics Testing Example

This example demonstrates how to access and verify the Prometheus metrics exposed by the CronJob Scale Down Operator.

## Prerequisites

1. CronJob Scale Down Operator deployed in your cluster
2. A CronJobScaleDown resource with some scaling/cleanup operations

## Accessing Metrics

### Method 1: Port Forward to Metrics Endpoint

```bash
# Port forward to the operator's metrics port
kubectl port-forward -n cronjob-scale-down-operator-system \
  deployment/cronjob-scale-down-operator-controller-manager 8443:8443
```

```bash
# Access metrics (in another terminal)
curl -k https://localhost:8443/metrics
```

### Method 2: Using a Debug Pod

```bash
# Create a debug pod to access metrics internally
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: metrics-debug
  namespace: cronjob-scale-down-operator-system
spec:
  containers:
  - name: curl
    image: curlimages/curl:latest
    command: ['sleep', '3600']
  restartPolicy: Never
EOF
```

```bash
# Execute curl from within the cluster
kubectl exec -n cronjob-scale-down-operator-system metrics-debug -- \
  curl -k https://cronjob-scale-down-operator-controller-manager-metrics-service:8443/metrics
```

## Expected Metrics

When accessing the metrics endpoint, you should see metrics like these:

### Scaling Operation Metrics
```prometheus
# HELP cronjob_scale_down_operations_total Total number of scale down operations performed
# TYPE cronjob_scale_down_operations_total counter
cronjob_scale_down_operations_total{name="my-scaler",namespace="default",target_kind="Deployment",target_name="my-app",target_namespace="default"} 5

# HELP cronjob_scale_up_operations_total Total number of scale up operations performed
# TYPE cronjob_scale_up_operations_total counter
cronjob_scale_up_operations_total{name="my-scaler",namespace="default",target_kind="Deployment",target_name="my-app",target_namespace="default"} 4

# HELP cronjob_scaled_down_resources_current Current number of resources that are scaled down
# TYPE cronjob_scaled_down_resources_current gauge
cronjob_scaled_down_resources_current{name="my-scaler",namespace="default",target_kind="Deployment",target_name="my-app",target_namespace="default"} 1
```

### Cleanup Operation Metrics
```prometheus
# HELP cronjob_cleanup_operations_total Total number of cleanup operations performed
# TYPE cronjob_cleanup_operations_total counter
cronjob_cleanup_operations_total{name="cleanup-job",namespace="default"} 3

# HELP cronjob_cleaned_resources_total Total number of resources cleaned up
# TYPE cronjob_cleaned_resources_total counter
cronjob_cleaned_resources_total{name="cleanup-job",namespace="default",resource_namespace="default",resource_type="ConfigMap"} 10
cronjob_cleaned_resources_total{name="cleanup-job",namespace="default",resource_namespace="default",resource_type="Secret"} 5
```

### Reconciliation Metrics
```prometheus
# HELP cronjob_reconciliations_total Total number of reconciliations performed
# TYPE cronjob_reconciliations_total counter
cronjob_reconciliations_total{name="my-scaler",namespace="default",result="success"} 25
cronjob_reconciliations_total{name="my-scaler",namespace="default",result="error"} 2

# HELP cronjob_reconciliation_duration_seconds Duration of reconciliation operations in seconds
# TYPE cronjob_reconciliation_duration_seconds histogram
cronjob_reconciliation_duration_seconds_bucket{name="my-scaler",namespace="default",le="0.005"} 15
cronjob_reconciliation_duration_seconds_bucket{name="my-scaler",namespace="default",le="0.01"} 20
cronjob_reconciliation_duration_seconds_bucket{name="my-scaler",namespace="default",le="0.025"} 25
```

### Resource Status Metrics
```prometheus
# HELP cronjob_target_resource_replicas_current Current number of replicas for target resources
# TYPE cronjob_target_resource_replicas_current gauge
cronjob_target_resource_replicas_current{name="my-scaler",namespace="default",replica_type="desired",target_kind="Deployment",target_name="my-app",target_namespace="default"} 0
cronjob_target_resource_replicas_current{name="my-scaler",namespace="default",replica_type="ready",target_kind="Deployment",target_name="my-app",target_namespace="default"} 0
cronjob_target_resource_replicas_current{name="my-scaler",namespace="default",replica_type="current",target_kind="Deployment",target_name="my-app",target_namespace="default"} 0
```

## Testing Scenarios

### 1. Verify Scaling Metrics

Create a test deployment and CronJobScaleDown:

```bash
# Create test deployment
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: metrics-test-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: metrics-test
  template:
    metadata:
      labels:
        app: metrics-test
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        resources:
          requests:
            memory: "16Mi"
            cpu: "10m"
          limits:
            memory: "32Mi"
            cpu: "20m"
EOF
```

```bash
# Create CronJobScaleDown for immediate testing
kubectl apply -f - <<EOF
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: metrics-test-scaler
  namespace: default
spec:
  targetRef:
    name: metrics-test-app
    namespace: default
    kind: Deployment
    apiVersion: apps/v1
  scaleDownSchedule: "*/30 * * * * *"  # Every 30 seconds for testing
  scaleUpSchedule: "*/60 * * * * *"    # Every 60 seconds for testing
  timeZone: "UTC"
EOF
```

Wait for a few minutes, then check metrics for:
- `cronjob_scale_down_operations_total`
- `cronjob_scale_up_operations_total`
- `cronjob_target_resource_replicas_current`

### 2. Verify Cleanup Metrics

Create test resources for cleanup:

```bash
# Create test ConfigMaps with cleanup annotations
for i in {1..5}; do
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config-$i
  namespace: default
  annotations:
    cleanup-after: "$(date -u -d '+1 minute' '+%Y-%m-%dT%H:%M:%SZ')"
data:
  test: "data-$i"
EOF
done
```

```bash
# Create cleanup-only CronJobScaleDown
kubectl apply -f - <<EOF
apiVersion: cronschedules.elbazi.co/v1
kind: CronJobScaleDown
metadata:
  name: metrics-cleanup-test
  namespace: default
spec:
  cleanupSchedule: "*/30 * * * * *"  # Every 30 seconds for testing
  cleanupConfig:
    annotationKey: "cleanup-after"
    resourceTypes: ["ConfigMap"]
    dryRun: false
  timeZone: "UTC"
EOF
```

Wait for cleanup to occur, then check metrics for:
- `cronjob_cleanup_operations_total`
- `cronjob_cleaned_resources_total`

### 3. Testing with Prometheus

If you have Prometheus running in your cluster:

```bash
# Query scaling operations rate
curl 'http://prometheus-server/api/v1/query?query=rate(cronjob_scale_down_operations_total[5m])'

# Query cleanup metrics
curl 'http://prometheus-server/api/v1/query?query=sum(cronjob_cleaned_resources_total)%20by%20(resource_type)'

# Query error rates
curl 'http://prometheus-server/api/v1/query?query=rate(cronjob_reconciliation_errors_total[5m])'
```

## Cleanup Test Resources

```bash
# Remove test resources
kubectl delete deployment metrics-test-app
kubectl delete cronjobscaledown metrics-test-scaler metrics-cleanup-test
kubectl delete configmap -l cleanup-after
kubectl delete pod metrics-debug -n cronjob-scale-down-operator-system
```

## Troubleshooting

### Metrics Not Available

1. **Check operator status:**
   ```bash
   kubectl get pods -n cronjob-scale-down-operator-system
   kubectl logs -n cronjob-scale-down-operator-system deployment/cronjob-scale-down-operator-controller-manager
   ```

2. **Verify metrics service:**
   ```bash
   kubectl get service -n cronjob-scale-down-operator-system cronjob-scale-down-operator-controller-manager-metrics-service
   ```

3. **Check RBAC permissions:**
   ```bash
   kubectl auth can-i get cronjobscaledowns --as=system:serviceaccount:cronjob-scale-down-operator-system:cronjob-scale-down-operator-controller-manager
   ```

### No Metrics Data

- Ensure CronJobScaleDown resources exist and are being reconciled
- Check if schedules are triggering (look at operator logs)
- Verify timezone configuration is correct

### Authentication Issues

The metrics endpoint uses HTTPS with authentication by default. Use `-k` flag with curl to skip certificate verification for testing.
