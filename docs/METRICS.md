# Prometheus Metrics

The CronJob Scale Down Operator exposes comprehensive Prometheus metrics to monitor scaling operations, cleanup activities, and overall system health.

## Metrics Categories

### Scaling Operation Metrics

#### `cronjob_scale_down_operations_total`
- **Type**: Counter
- **Description**: Total number of scale down operations performed
- **Labels**: 
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `target_kind`: Target resource kind (Deployment, StatefulSet)
  - `target_name`: Target resource name
  - `target_namespace`: Target resource namespace

#### `cronjob_scale_up_operations_total`
- **Type**: Counter
- **Description**: Total number of scale up operations performed
- **Labels**: Same as scale down operations

#### `cronjob_scaled_down_resources_current`
- **Type**: Gauge
- **Description**: Current number of resources that are scaled down (1 = scaled down, 0 = not scaled down)
- **Labels**: Same as scale operations

#### `cronjob_scale_operation_duration_seconds`
- **Type**: Histogram
- **Description**: Duration of scale operations in seconds
- **Labels**:
  - `operation_type`: Type of operation (scale_down, scale_up)
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name

### Cleanup Operation Metrics

#### `cronjob_cleanup_operations_total`
- **Type**: Counter
- **Description**: Total number of cleanup operations performed
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name

#### `cronjob_cleaned_resources_total`
- **Type**: Counter
- **Description**: Total number of resources cleaned up
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `resource_type`: Type of cleaned resource (ConfigMap, Secret, Pod, etc.)
  - `resource_namespace`: Namespace of cleaned resource

#### `cronjob_cleaned_resources_by_label_total`
- **Type**: Counter
- **Description**: Total number of resources cleaned up by label selectors
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `resource_type`: Type of cleaned resource
  - `label_key`: Label key used for selection
  - `label_value`: Label value used for selection

#### `cronjob_cleanup_operation_duration_seconds`
- **Type**: Histogram
- **Description**: Duration of cleanup operations in seconds
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name

#### `cronjob_orphan_resources_cleaned_total`
- **Type**: Counter
- **Description**: Total number of orphan resources cleaned up (resources without cleanup annotations)
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `resource_type`: Type of cleaned resource
  - `resource_namespace`: Namespace of cleaned resource

#### `cronjob_dry_run_operations_total`
- **Type**: Counter
- **Description**: Total number of dry run operations performed
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `operation_type`: Type of operation (cleanup)

### Reconciliation Metrics

#### `cronjob_reconciliations_total`
- **Type**: Counter
- **Description**: Total number of reconciliations performed
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `result`: Result of reconciliation (success, error)

#### `cronjob_reconciliation_duration_seconds`
- **Type**: Histogram
- **Description**: Duration of reconciliation operations in seconds
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name

#### `cronjob_reconciliation_errors_total`
- **Type**: Counter
- **Description**: Total number of reconciliation errors
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `error_type`: Type of error (fetch_error, scale_down_error, scale_up_error, cleanup_error, status_update_error)

### Resource Status Metrics

#### `cronjob_active_resources_current`
- **Type**: Gauge
- **Description**: Current number of active CronJobScaleDown resources
- **Labels**:
  - `namespace`: Namespace containing the resources

#### `cronjob_target_resource_replicas_current`
- **Type**: Gauge
- **Description**: Current number of replicas for target resources
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `target_kind`: Target resource kind
  - `target_name`: Target resource name
  - `target_namespace`: Target resource namespace
  - `replica_type`: Type of replica count (desired, ready, current)

### Schedule Metrics

#### `cronjob_schedule_next_execution_timestamp`
- **Type**: Gauge
- **Description**: Unix timestamp of the next scheduled execution
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `schedule_type`: Type of schedule (scale_down, scale_up, cleanup)

#### `cronjob_schedule_last_execution_timestamp`
- **Type**: Gauge
- **Description**: Unix timestamp of the last execution
- **Labels**: Same as next execution timestamp

### Configuration Metrics

#### `cronjob_configuration_validation_errors_total`
- **Type**: Counter
- **Description**: Total number of configuration validation errors
- **Labels**:
  - `namespace`: CronJobScaleDown resource namespace
  - `name`: CronJobScaleDown resource name
  - `validation_type`: Type of validation error (spec_validation, timezone_validation, schedule_validation)

## Example PromQL Queries

### Monitor Scaling Operations
```promql
# Rate of scale down operations
rate(cronjob_scale_down_operations_total[5m])

# Rate of scale up operations
rate(cronjob_scale_up_operations_total[5m])

# Currently scaled down resources
sum(cronjob_scaled_down_resources_current) by (namespace, target_kind)
```

### Monitor Cleanup Operations
```promql
# Rate of cleanup operations
rate(cronjob_cleanup_operations_total[5m])

# Resources cleaned by type
sum(rate(cronjob_cleaned_resources_total[5m])) by (resource_type)

# Orphan resources cleaned
sum(rate(cronjob_orphan_resources_cleaned_total[5m])) by (resource_type)
```

### Monitor System Health
```promql
# Reconciliation error rate
rate(cronjob_reconciliation_errors_total[5m])

# Configuration validation errors
rate(cronjob_configuration_validation_errors_total[5m])

# Active CronJobScaleDown resources
sum(cronjob_active_resources_current) by (namespace)
```

### Monitor Resource Status
```promql
# Replica status by resource
cronjob_target_resource_replicas_current{replica_type="ready"} / cronjob_target_resource_replicas_current{replica_type="desired"}

# Resources with no ready replicas
cronjob_target_resource_replicas_current{replica_type="ready"} == 0
```

### Monitor Schedules
```promql
# Time until next execution
cronjob_schedule_next_execution_timestamp - time()

# Time since last execution
time() - cronjob_schedule_last_execution_timestamp
```

## Grafana Dashboard Queries

### Panel: Scale Operations Rate
```promql
sum(rate(cronjob_scale_down_operations_total[5m])) by (namespace, name)
sum(rate(cronjob_scale_up_operations_total[5m])) by (namespace, name)
```

### Panel: Cleanup Operations
```promql
sum(rate(cronjob_cleaned_resources_total[5m])) by (resource_type, namespace)
```

### Panel: Error Rate
```promql
sum(rate(cronjob_reconciliation_errors_total[5m])) by (error_type, namespace)
```

### Panel: Active Resources
```promql
sum(cronjob_active_resources_current) by (namespace)
```

## Alerts

### High Error Rate
```yaml
- alert: CronJobScaleDownHighErrorRate
  expr: rate(cronjob_reconciliation_errors_total[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High error rate in CronJob Scale Down operator"
    description: "Error rate is {{ $value }} errors per second for {{ $labels.namespace }}/{{ $labels.name }}"
```

### Scale Operation Failures
```yaml
- alert: CronJobScaleOperationStuck
  expr: time() - cronjob_schedule_last_execution_timestamp{schedule_type!=""} > 3600
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "CronJob scale operation appears stuck"
    description: "No {{ $labels.schedule_type }} execution for {{ $labels.namespace }}/{{ $labels.name }} in over 1 hour"
```

### Configuration Issues
```yaml
- alert: CronJobConfigurationErrors
  expr: increase(cronjob_configuration_validation_errors_total[1h]) > 0
  for: 0m
  labels:
    severity: warning
  annotations:
    summary: "CronJob configuration validation errors detected"
    description: "{{ $value }} configuration errors for {{ $labels.namespace }}/{{ $labels.name }}"
```

## Accessing Metrics

The metrics are exposed on the `/metrics` endpoint of the operator. By default, this is available on port `8443` with HTTPS and authentication enabled.

### Using kubectl port-forward
```bash
kubectl port-forward -n cronjob-scale-down-operator-system deployment/cronjob-scale-down-operator-controller-manager 8443:8443
curl -k https://localhost:8443/metrics
```

### ServiceMonitor for Prometheus Operator
The operator includes a ServiceMonitor resource for automatic discovery by Prometheus Operator. This is located in `config/prometheus/monitor.yaml`.

### Manual Prometheus Configuration
```yaml
scrape_configs:
  - job_name: 'cronjob-scale-down-operator'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
            - cronjob-scale-down-operator-system
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: cronjob-scale-down-operator-controller-manager-metrics-service
    scheme: https
    tls_config:
      insecure_skip_verify: true
```

## Metric Labels Best Practices

1. **Consistent Labeling**: All metrics use consistent label naming (namespace, name for CronJobScaleDown resources)
2. **Cardinality Control**: Labels are designed to avoid high cardinality issues
3. **Meaningful Labels**: Labels provide context for filtering and aggregation
4. **Resource Identification**: Target resource information is included where relevant
5. **Operation Context**: Operation types and error types are clearly labeled

## Integration with Monitoring Stack

### Prometheus
- Metrics are automatically discovered via ServiceMonitor
- Built-in alerting rules can be configured
- Historical data retention for trend analysis

### Grafana
- Pre-built dashboard queries provided
- Visualization of scaling patterns and cleanup activities
- Alerting integration for operational visibility

### AlertManager
- Sample alert rules for common failure scenarios
- Integration with notification channels (Slack, email, etc.)
- Escalation policies for critical issues
