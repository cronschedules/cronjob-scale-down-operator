/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// Scaling operation metrics
	ScaleDownOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_scale_down_operations_total",
			Help: "Total number of scale down operations performed",
		},
		[]string{"namespace", "name", "target_kind", "target_name", "target_namespace"},
	)

	ScaleUpOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_scale_up_operations_total",
			Help: "Total number of scale up operations performed",
		},
		[]string{"namespace", "name", "target_kind", "target_name", "target_namespace"},
	)

	CurrentScaledDownResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cronjob_scaled_down_resources_current",
			Help: "Current number of resources that are scaled down",
		},
		[]string{"namespace", "name", "target_kind", "target_name", "target_namespace"},
	)

	ScaleOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cronjob_scale_operation_duration_seconds",
			Help:    "Duration of scale operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation_type", "namespace", "name"},
	)

	// Cleanup operation metrics
	CleanupOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_cleanup_operations_total",
			Help: "Total number of cleanup operations performed",
		},
		[]string{"namespace", "name"},
	)

	CleanedResourcesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_cleaned_resources_total",
			Help: "Total number of resources cleaned up",
		},
		[]string{"namespace", "name", "resource_type", "resource_namespace"},
	)

	CleanedResourcesByLabelTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_cleaned_resources_by_label_total",
			Help: "Total number of resources cleaned up by label selectors",
		},
		[]string{"namespace", "name", "resource_type", "label_key", "label_value"},
	)

	CleanupOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cronjob_cleanup_operation_duration_seconds",
			Help:    "Duration of cleanup operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "name"},
	)

	OrphanResourcesCleanedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_orphan_resources_cleaned_total",
			Help: "Total number of orphan resources cleaned up (resources without cleanup annotations)",
		},
		[]string{"namespace", "name", "resource_type", "resource_namespace"},
	)

	// Reconciliation metrics
	ReconciliationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_reconciliations_total",
			Help: "Total number of reconciliations performed",
		},
		[]string{"namespace", "name", "result"},
	)

	ReconciliationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cronjob_reconciliation_duration_seconds",
			Help:    "Duration of reconciliation operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "name"},
	)

	ReconciliationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_reconciliation_errors_total",
			Help: "Total number of reconciliation errors",
		},
		[]string{"namespace", "name", "error_type"},
	)

	// Resource status metrics
	ActiveCronJobScaleDownResources = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cronjob_active_resources_current",
			Help: "Current number of active CronJobScaleDown resources",
		},
		[]string{"namespace"},
	)

	TargetResourceReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cronjob_target_resource_replicas_current",
			Help: "Current number of replicas for target resources",
		},
		[]string{"namespace", "name", "target_kind", "target_name", "target_namespace", "replica_type"},
	)

	// Schedule metrics
	ScheduleExecutionTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cronjob_schedule_next_execution_timestamp",
			Help: "Unix timestamp of the next scheduled execution",
		},
		[]string{"namespace", "name", "schedule_type"},
	)

	LastExecutionTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cronjob_schedule_last_execution_timestamp",
			Help: "Unix timestamp of the last execution",
		},
		[]string{"namespace", "name", "schedule_type"},
	)

	// Configuration metrics
	ConfigurationValidationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_configuration_validation_errors_total",
			Help: "Total number of configuration validation errors",
		},
		[]string{"namespace", "name", "validation_type"},
	)

	DryRunOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_dry_run_operations_total",
			Help: "Total number of dry run operations performed",
		},
		[]string{"namespace", "name", "operation_type"},
	)
)

// init registers all metrics with the controller-runtime metrics registry
func init() {
	// Register scaling metrics
	metrics.Registry.MustRegister(
		ScaleDownOperationsTotal,
		ScaleUpOperationsTotal,
		CurrentScaledDownResources,
		ScaleOperationDuration,
	)

	// Register cleanup metrics
	metrics.Registry.MustRegister(
		CleanupOperationsTotal,
		CleanedResourcesTotal,
		CleanedResourcesByLabelTotal,
		CleanupOperationDuration,
		OrphanResourcesCleanedTotal,
		DryRunOperationsTotal,
	)

	// Register reconciliation metrics
	metrics.Registry.MustRegister(
		ReconciliationTotal,
		ReconciliationDuration,
		ReconciliationErrors,
	)

	// Register resource status metrics
	metrics.Registry.MustRegister(
		ActiveCronJobScaleDownResources,
		TargetResourceReplicas,
	)

	// Register schedule metrics
	metrics.Registry.MustRegister(
		ScheduleExecutionTime,
		LastExecutionTime,
	)

	// Register configuration metrics
	metrics.Registry.MustRegister(
		ConfigurationValidationErrors,
	)
}

// RecordScaleDownOperation records a scale down operation
func RecordScaleDownOperation(namespace, name, targetKind, targetName, targetNamespace string) {
	ScaleDownOperationsTotal.WithLabelValues(namespace, name, targetKind, targetName, targetNamespace).Inc()
	CurrentScaledDownResources.WithLabelValues(namespace, name, targetKind, targetName, targetNamespace).Set(1)
}

// RecordScaleUpOperation records a scale up operation
func RecordScaleUpOperation(namespace, name, targetKind, targetName, targetNamespace string) {
	ScaleUpOperationsTotal.WithLabelValues(namespace, name, targetKind, targetName, targetNamespace).Inc()
	CurrentScaledDownResources.WithLabelValues(namespace, name, targetKind, targetName, targetNamespace).Set(0)
}

// RecordCleanupOperation records a cleanup operation
func RecordCleanupOperation(namespace, name string, duration float64) {
	CleanupOperationsTotal.WithLabelValues(namespace, name).Inc()
	CleanupOperationDuration.WithLabelValues(namespace, name).Observe(duration)
}

// RecordCleanedResource records a cleaned resource
func RecordCleanedResource(namespace, name, resourceType, resourceNamespace string) {
	CleanedResourcesTotal.WithLabelValues(namespace, name, resourceType, resourceNamespace).Inc()
}

// RecordCleanedResourceByLabel records a cleaned resource with label information
func RecordCleanedResourceByLabel(namespace, name, resourceType, labelKey, labelValue string) {
	CleanedResourcesByLabelTotal.WithLabelValues(namespace, name, resourceType, labelKey, labelValue).Inc()
}

// RecordOrphanResourceCleaned records an orphan resource cleanup
func RecordOrphanResourceCleaned(namespace, name, resourceType, resourceNamespace string) {
	OrphanResourcesCleanedTotal.WithLabelValues(namespace, name, resourceType, resourceNamespace).Inc()
}

// RecordReconciliation records a reconciliation operation
func RecordReconciliation(namespace, name, result string, duration float64) {
	ReconciliationTotal.WithLabelValues(namespace, name, result).Inc()
	ReconciliationDuration.WithLabelValues(namespace, name).Observe(duration)
}

// RecordReconciliationError records a reconciliation error
func RecordReconciliationError(namespace, name, errorType string) {
	ReconciliationErrors.WithLabelValues(namespace, name, errorType).Inc()
	ReconciliationTotal.WithLabelValues(namespace, name, "error").Inc()
}

// RecordConfigurationValidationError records a configuration validation error
func RecordConfigurationValidationError(namespace, name, validationType string) {
	ConfigurationValidationErrors.WithLabelValues(namespace, name, validationType).Inc()
}

// RecordDryRunOperation records a dry run operation
func RecordDryRunOperation(namespace, name, operationType string) {
	DryRunOperationsTotal.WithLabelValues(namespace, name, operationType).Inc()
}

// UpdateActiveResources updates the count of active CronJobScaleDown resources
func UpdateActiveResources(namespace string, count float64) {
	ActiveCronJobScaleDownResources.WithLabelValues(namespace).Set(count)
}

// UpdateTargetResourceReplicas updates the replica count for target resources
func UpdateTargetResourceReplicas(namespace, name, targetKind, targetName, targetNamespace, replicaType string, count float64) {
	TargetResourceReplicas.WithLabelValues(namespace, name, targetKind, targetName, targetNamespace, replicaType).Set(count)
}

// UpdateScheduleExecutionTime updates the next execution time for a schedule
func UpdateScheduleExecutionTime(namespace, name, scheduleType string, timestamp float64) {
	ScheduleExecutionTime.WithLabelValues(namespace, name, scheduleType).Set(timestamp)
}

// UpdateLastExecutionTime updates the last execution time for a schedule
func UpdateLastExecutionTime(namespace, name, scheduleType string, timestamp float64) {
	LastExecutionTime.WithLabelValues(namespace, name, scheduleType).Set(timestamp)
}

// ResetResourceMetrics resets metrics for a deleted resource
func ResetResourceMetrics(namespace, name, targetKind, targetName, targetNamespace string) {
	// Reset scaling metrics
	CurrentScaledDownResources.DeleteLabelValues(namespace, name, targetKind, targetName, targetNamespace)

	// Reset target resource metrics
	TargetResourceReplicas.DeleteLabelValues(namespace, name, targetKind, targetName, targetNamespace, "desired")
	TargetResourceReplicas.DeleteLabelValues(namespace, name, targetKind, targetName, targetNamespace, "ready")
	TargetResourceReplicas.DeleteLabelValues(namespace, name, targetKind, targetName, targetNamespace, "current")

	// Reset schedule metrics
	ScheduleExecutionTime.DeleteLabelValues(namespace, name, "scale_down")
	ScheduleExecutionTime.DeleteLabelValues(namespace, name, "scale_up")
	ScheduleExecutionTime.DeleteLabelValues(namespace, name, "cleanup")

	LastExecutionTime.DeleteLabelValues(namespace, name, "scale_down")
	LastExecutionTime.DeleteLabelValues(namespace, name, "scale_up")
	LastExecutionTime.DeleteLabelValues(namespace, name, "cleanup")
}
