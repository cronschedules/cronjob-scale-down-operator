package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
	"github.com/z4ck404/cronjob-scale-down-operator/internal/metrics"
)

// K8sClient wraps a kubernetes client
type K8sClient struct {
	client.Client
}

type TargetObject struct {
	cronschedulesv1.TargetRef
}

// retryOnConflict retries the given function if it encounters a conflict error
func (c *K8sClient) retryOnConflict(ctx context.Context, retryFn func() error) error {
	return wait.ExponentialBackoffWithContext(ctx, wait.Backoff{
		Duration: 100 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
		Steps:    5, // Retry up to 5 times
	}, func(context.Context) (bool, error) {
		err := retryFn()
		if err == nil {
			return true, nil // Success, stop retrying
		}
		if apierrors.IsConflict(err) {
			return false, nil // Retry on conflict
		}
		return false, err // Stop retrying on other errors
	})
}

const (
	annotationKeyOriginalReplicas = "cronjob-scale-down-operator/original-replicas"
	DeploymentKind                = "Deployment"
	StatefulSetKind               = "StatefulSet"
)

// Documentation of the logic:
// 1. Get the target resource (deployment or statefulset)
// 2. Scale down the target resource to 0 replicas
// 3. Update the target resource status with the last scale down time
// 4. If the next execution time is in the future, return and wait for the next execution time

//lint:ignore U1000 Ignore unused function
func scaleDown(ctx context.Context, c *K8sClient, targetRef TargetObject) error {
	logger := log.FromContext(ctx)
	logger.Info("Scaling down the target resource", "targetRef", targetRef)

	// Scale down the target resource
	err := c.ScaleDownTargetResource(ctx, targetRef)
	if err != nil {
		logger.Error(err, "Error scaling down target resource")
		return err
	}

	return nil
}

func (c *K8sClient) ScaleDownTargetResource(ctx context.Context, targetRef TargetObject) error {
	logger := log.FromContext(ctx)

	switch targetRef.Kind {
	case DeploymentKind:
		deployment := &appsv1.Deployment{}
		err := c.Get(ctx, client.ObjectKey{Name: targetRef.Name, Namespace: targetRef.Namespace}, deployment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Deployment not found, skipping scale down", "name", targetRef.Name, "namespace", targetRef.Namespace)
				return nil
			}
			logger.Error(err, "Error getting deployment from the cluster", "name", targetRef.Name)
			return err
		}

		// Ensure original replicas annotation is set before scaling down
		if err := c.UpdateTargetResourceOriginalReplicasAnnotation(ctx, targetRef); err != nil {
			logger.Error(err, "Failed to set original replicas annotation before scaling down")
			return err
		}

		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
			logger.Info("Deployment is already scaled down, skipping", "name", deployment.GetName())
			return nil
		}

		err = c.scaleDownDeployment(ctx, deployment)
		if err != nil {
			logger.Error(err, "Error scaling down deployment", "name", deployment.GetName())
			return err
		}

		logger.Info("Deployment scaled down successfully", "name", deployment.GetName())

	case StatefulSetKind:
		statefulset := &appsv1.StatefulSet{}
		err := c.Get(ctx, client.ObjectKey{Name: targetRef.Name, Namespace: targetRef.Namespace}, statefulset)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("StatefulSet not found, skipping scale down", "name", targetRef.Name, "namespace", targetRef.Namespace)
				return nil
			}
			logger.Error(err, "Error getting statefulset from the cluster", "name", targetRef.Name)
			return err
		}

		// Ensure original replicas annotation is set before scaling down
		if err := c.UpdateTargetResourceOriginalReplicasAnnotation(ctx, targetRef); err != nil {
			logger.Error(err, "Failed to set original replicas annotation before scaling down")
			return err
		}

		if statefulset.Spec.Replicas != nil && *statefulset.Spec.Replicas == 0 {
			logger.Info("Statefulset is already scaled down, skipping", "name", statefulset.GetName())
			return nil
		}

		err = c.scaleDownStatefulset(ctx, statefulset)
		if err != nil {
			logger.Error(err, "Error scaling down statefulset", "name", statefulset.GetName())
			return err
		}

		logger.Info("Statefulset scaled down successfully", "name", statefulset.GetName())
	default:
		logger.Error(nil, "Unsupported target resource kind", "kind", targetRef.Kind)
		return fmt.Errorf("unsupported target resource kind: %s", targetRef.Kind)
	}

	return nil
}

//lint:ignore U1000 Ignore unused function
func scaleUpTargetResource(ctx context.Context, targetResource client.Object) error {

	return nil
}

func (c *K8sClient) scaleDownDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	logger := log.FromContext(ctx)

	return c.retryOnConflict(ctx, func() error {
		// Get fresh copy of the deployment to avoid conflicts
		fresh := &appsv1.Deployment{}
		if err := c.Get(ctx, client.ObjectKeyFromObject(deployment), fresh); err != nil {
			return err
		}

		// Set replicas to 0
		fresh.Spec.Replicas = ptr.To[int32](0)

		// Update the deployment
		if err := c.Update(ctx, fresh); err != nil {
			logger.V(1).Info("Retry needed for deployment scale down", "name", fresh.GetName(), "error", err.Error())
			return err
		}

		logger.Info("Successfully scaled down deployment", "name", fresh.GetName())
		return nil
	})
}

func (c *K8sClient) scaleDownStatefulset(ctx context.Context, statefulset *appsv1.StatefulSet) error {
	logger := log.FromContext(ctx)

	return c.retryOnConflict(ctx, func() error {
		// Get fresh copy of the statefulset to avoid conflicts
		fresh := &appsv1.StatefulSet{}
		if err := c.Get(ctx, client.ObjectKeyFromObject(statefulset), fresh); err != nil {
			return err
		}

		// Set replicas to 0
		fresh.Spec.Replicas = ptr.To[int32](0)

		// Update the statefulset
		if err := c.Update(ctx, fresh); err != nil {
			logger.V(1).Info("Retry needed for statefulset scale down", "name", fresh.GetName(), "error", err.Error())
			return err
		}

		logger.Info("Successfully scaled down statefulset", "name", fresh.GetName())
		return nil
	})
}

func (c *K8sClient) GetReplicasCount(ctx context.Context, targetResource TargetObject) *int32 {
	logger := log.FromContext(ctx)
	var replicas *int32

	switch targetResource.Kind {
	case DeploymentKind:
		deployment := &appsv1.Deployment{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, deployment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Deployment not found for replica count check", "name", targetResource.Name, "namespace", targetResource.Namespace)
				return nil
			}
			logger.Error(err, "Error getting deployment from the cluster", "name", targetResource.Name)
			return nil
		}
		replicas = deployment.Spec.Replicas

	case StatefulSetKind:
		statefulset := &appsv1.StatefulSet{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, statefulset)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("StatefulSet not found for replica count check", "name", targetResource.Name, "namespace", targetResource.Namespace)
				return nil
			}
			logger.Error(err, "Error getting statefulset from the cluster", "name", targetResource.Name)
			return nil
		}
		replicas = statefulset.Spec.Replicas
	default:
		logger.Error(nil, "Unsupported target resource kind", "kind", targetResource.Kind)
	}

	return replicas
}

func (c *K8sClient) UpdateTargetResourceOriginalReplicasAnnotation(ctx context.Context, targetResource TargetObject) error {
	logger := log.FromContext(ctx)
	var targetResourceObject client.Object

	switch targetResource.Kind {
	case DeploymentKind:
		targetResourceObject = &appsv1.Deployment{}
	case StatefulSetKind:
		targetResourceObject = &appsv1.StatefulSet{}
	default:
		logger.Error(nil, "Unsupported target resource kind for annotation", "kind", targetResource.Kind)
		return fmt.Errorf("unsupported target resource kind: %s", targetResource.Kind)
	}

	return c.retryOnConflict(ctx, func() error {
		// Get fresh copy of the resource
		if err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, targetResourceObject); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Target resource not found for annotation update", "name", targetResource.Name, "namespace", targetResource.Namespace, "kind", targetResource.Kind)
				return nil
			}
			logger.Error(err, "Failed to get target resource for annotation", "name", targetResource.Name)
			return fmt.Errorf("failed to get target resource: %w", err)
		}

		annotations := targetResourceObject.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		if _, ok := annotations[annotationKeyOriginalReplicas]; ok {
			logger.Info("Original replicas annotation already exists", "name", targetResource.Name)
			return nil
		}

		originalTargetResourceReplicas := c.GetReplicasCount(ctx, targetResource)
		if originalTargetResourceReplicas == nil {
			logger.Error(nil, "Failed to get original replicas count for target resource", "name", targetResource.Name)
			return fmt.Errorf("failed to get original replicas count for target resource")
		}

		annotations[annotationKeyOriginalReplicas] = strconv.Itoa(int(*originalTargetResourceReplicas))
		targetResourceObject.SetAnnotations(annotations)

		if err := c.Update(ctx, targetResourceObject); err != nil {
			logger.V(1).Info("Retry needed for annotation update", "name", targetResource.Name, "error", err.Error())
			return err
		}

		logger.Info("Successfully updated original replicas annotation", "name", targetResource.Name, "replicas", *originalTargetResourceReplicas)
		return nil
	})
}

// ScaleUpTargetResource scales up the target resource to its original replica count (from annotation)
func (c *K8sClient) ScaleUpTargetResource(ctx context.Context, targetRef TargetObject) error {
	logger := log.FromContext(ctx)

	return c.retryOnConflict(ctx, func() error {
		var obj client.Object

		switch targetRef.Kind {
		case DeploymentKind:
			obj = &appsv1.Deployment{}
		case StatefulSetKind:
			obj = &appsv1.StatefulSet{}
		default:
			logger.Error(nil, "Unsupported target resource kind for scale up", "kind", targetRef.Kind)
			return fmt.Errorf("unsupported target resource kind: %s", targetRef.Kind)
		}

		// Get fresh copy of the resource
		if err := c.Get(ctx, client.ObjectKey{Name: targetRef.Name, Namespace: targetRef.Namespace}, obj); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Target resource not found, skipping scale up", "name", targetRef.Name, "namespace", targetRef.Namespace, "kind", targetRef.Kind)
				return nil
			}
			logger.Error(err, "Failed to get target resource for scale up", "name", targetRef.Name)
			return err
		}

		annotations := obj.GetAnnotations()
		if annotations == nil {
			logger.Error(nil, "No annotations found on target resource for scale up", "name", targetRef.Name)
			return fmt.Errorf("no annotations found on target resource")
		}
		val, ok := annotations[annotationKeyOriginalReplicas]
		if !ok {
			logger.Error(nil, "Original replicas annotation not found for scale up", "name", targetRef.Name)
			return fmt.Errorf("original replicas annotation not found")
		}
		originalReplicas, err := strconv.Atoi(val)
		if err != nil {
			logger.Error(err, "Invalid original replicas annotation value", "value", val)
			return err
		}

		switch o := obj.(type) {
		case *appsv1.Deployment:
			o.Spec.Replicas = ptr.To[int32](int32(originalReplicas))
			if err := c.Update(ctx, o); err != nil {
				logger.V(1).Info("Retry needed for deployment scale up", "name", o.GetName(), "error", err.Error())
				return err
			}
			logger.Info("Successfully scaled up deployment", "name", o.GetName(), "replicas", originalReplicas)
		case *appsv1.StatefulSet:
			o.Spec.Replicas = ptr.To[int32](int32(originalReplicas))
			if err := c.Update(ctx, o); err != nil {
				logger.V(1).Info("Retry needed for statefulset scale up", "name", o.GetName(), "error", err.Error())
				return err
			}
			logger.Info("Successfully scaled up statefulset", "name", o.GetName(), "replicas", originalReplicas)
		default:
			logger.Error(nil, "Unsupported resource type for scaling", "type", fmt.Sprintf("%T", obj))
			return fmt.Errorf("unsupported resource type: %T", obj)
		}

		return nil
	})
}

// CleanupResourcesWithMetrics finds and deletes resources based on cleanup configuration and tracks metrics
func (c *K8sClient) CleanupResourcesWithMetrics(ctx context.Context, cleanupConfig *cronschedulesv1.CleanupConfig, defaultNamespace, cronJobNamespace, cronJobName string) (int32, error) {
	logger := log.FromContext(ctx)

	if cleanupConfig == nil {
		return 0, fmt.Errorf("cleanup config is nil")
	}

	// Use default namespace if none specified
	namespaces := cleanupConfig.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{defaultNamespace}
	}

	var totalDeleted int32

	for _, resourceType := range cleanupConfig.ResourceTypes {
		for _, namespace := range namespaces {
			deleted, err := c.cleanupResourceTypeWithMetrics(ctx, resourceType, namespace, cleanupConfig, cronJobNamespace, cronJobName)
			if err != nil {
				logger.Error(err, "Failed to cleanup resource type", "type", resourceType, "namespace", namespace)
				continue
			}
			totalDeleted += deleted
		}
	}

	logger.Info("Cleanup operation completed", "totalDeleted", totalDeleted, "dryRun", cleanupConfig.DryRun)
	return totalDeleted, nil
}

// cleanupResourceTypeWithMetrics handles cleanup for a specific resource type in a namespace with metrics tracking
func (c *K8sClient) cleanupResourceTypeWithMetrics(ctx context.Context, resourceType, namespace string, cleanupConfig *cronschedulesv1.CleanupConfig, cronJobNamespace, cronJobName string) (int32, error) {
	objList, err := c.createResourceList(resourceType)
	if err != nil {
		return 0, err
	}

	listOpts := c.buildListOptions(resourceType, namespace, cleanupConfig)

	// List resources
	if err := c.List(ctx, objList, listOpts...); err != nil {
		return 0, fmt.Errorf("failed to list %s in namespace %s: %w", resourceType, namespace, err)
	}

	return c.processResourceListWithMetrics(ctx, objList, cleanupConfig, resourceType, namespace, cronJobNamespace, cronJobName), nil
}

// createResourceList creates the appropriate list object for the resource type
func (c *K8sClient) createResourceList(resourceType string) (client.ObjectList, error) {
	switch resourceType {
	case "Deployment":
		return &appsv1.DeploymentList{}, nil
	case "StatefulSet":
		return &appsv1.StatefulSetList{}, nil
	case "Service":
		return &corev1.ServiceList{}, nil
	case "ConfigMap":
		return &corev1.ConfigMapList{}, nil
	case "Secret":
		return &corev1.SecretList{}, nil
	case "Pod":
		return &corev1.PodList{}, nil
	case "Job":
		return &batchv1.JobList{}, nil
	case "Role":
		return &rbacv1.RoleList{}, nil
	case "RoleBinding":
		return &rbacv1.RoleBindingList{}, nil
	case "ClusterRole":
		return &rbacv1.ClusterRoleList{}, nil
	case "ClusterRoleBinding":
		return &rbacv1.ClusterRoleBindingList{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// buildListOptions builds the list options for querying resources
func (c *K8sClient) buildListOptions(resourceType, namespace string, cleanupConfig *cronschedulesv1.CleanupConfig) []client.ListOption {
	listOpts := []client.ListOption{}

	// Namespace scoping: ClusterRole and ClusterRoleBinding are cluster-scoped
	if resourceType != "ClusterRole" && resourceType != "ClusterRoleBinding" {
		listOpts = append(listOpts, client.InNamespace(namespace))
	}

	// Add label selector if specified
	if len(cleanupConfig.LabelSelector) > 0 {
		selector := labels.SelectorFromSet(cleanupConfig.LabelSelector)
		listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: selector})
	}

	return listOpts
}

// processResourceListWithMetrics processes the list of resources for cleanup with metrics tracking
func (c *K8sClient) processResourceListWithMetrics(ctx context.Context, objList client.ObjectList, cleanupConfig *cronschedulesv1.CleanupConfig, resourceType, resourceNamespace, cronJobNamespace, cronJobName string) int32 {
	var deletedCount int32
	items := c.extractItemsFromList(objList)

	for _, obj := range items {
		if c.shouldCleanupResource(ctx, obj, cleanupConfig) {
			deletedCount += c.deleteResourceWithMetrics(ctx, obj, cleanupConfig.DryRun, resourceType, resourceNamespace, cronJobNamespace, cronJobName)

			// Check if this is an orphan resource (no cleanup annotation)
			if cleanupConfig.CleanupOrphanResources && c.isOrphanResourceForCleanup(ctx, obj, cleanupConfig) {
				metrics.RecordOrphanResourceCleaned(cronJobNamespace, cronJobName, resourceType, resourceNamespace)
			}
		}
	}

	return deletedCount
}

// extractItemsFromList extracts items from different list types
func (c *K8sClient) extractItemsFromList(objList client.ObjectList) []client.Object {
	var items []client.Object

	switch list := objList.(type) {
	case *appsv1.DeploymentList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *appsv1.StatefulSetList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *corev1.ServiceList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *corev1.ConfigMapList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *corev1.SecretList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *corev1.PodList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *batchv1.JobList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *rbacv1.RoleList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *rbacv1.RoleBindingList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *rbacv1.ClusterRoleList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	case *rbacv1.ClusterRoleBindingList:
		for i := range list.Items {
			items = append(items, &list.Items[i])
		}
	}

	return items
}

// deleteResourceWithMetrics handles the actual deletion or dry-run logging with metrics
func (c *K8sClient) deleteResourceWithMetrics(ctx context.Context, obj client.Object, dryRun bool, resourceType, resourceNamespace, cronJobNamespace, cronJobName string) int32 {
	logger := log.FromContext(ctx)

	if dryRun {
		logger.Info("DRY RUN: Would delete resource",
			"type", resourceType,
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())
		return 1
	}

	if err := c.Delete(ctx, obj); err != nil {
		logger.Error(err, "Failed to delete resource",
			"type", resourceType,
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())
		return 0
	}

	logger.Info("Resource deleted successfully",
		"type", resourceType,
		"name", obj.GetName(),
		"namespace", obj.GetNamespace())

	// Record metrics for cleaned resource
	metrics.RecordCleanedResource(cronJobNamespace, cronJobName, resourceType, resourceNamespace)

	// Record metrics for cleaned resource by label if labels exist
	for labelKey, labelValue := range obj.GetLabels() {
		metrics.RecordCleanedResourceByLabel(cronJobNamespace, cronJobName, resourceType, labelKey, labelValue)
	}

	return 1
}

// shouldCleanupResource determines if a resource should be cleaned up based on annotations or orphan rules
func (c *K8sClient) shouldCleanupResource(ctx context.Context, obj client.Object, cleanupConfig *cronschedulesv1.CleanupConfig) bool {
	logger := log.FromContext(ctx)

	annotations := obj.GetAnnotations()

	// Check if cleanup annotation exists
	cleanupValue, hasCleanupAnnotation := "", false
	if annotations != nil {
		cleanupValue, hasCleanupAnnotation = annotations[cleanupConfig.AnnotationKey]
	}

	// Handle annotated resources (existing logic)
	if hasCleanupAnnotation {
		// If annotation value is empty, clean up immediately
		if cleanupValue == "" {
			logger.Info("Resource marked for immediate cleanup",
				"name", obj.GetName(),
				"namespace", obj.GetNamespace())
			return true
		}

		// Parse cleanup time/duration
		return c.isCleanupTimeReached(ctx, cleanupValue, obj)
	}

	// Handle orphan resources (new logic)
	if cleanupConfig.CleanupOrphanResources {
		return c.isOrphanResourceForCleanup(ctx, obj, cleanupConfig)
	}

	// Resource has no cleanup annotation and orphan cleanup is disabled
	return false
}

// isOrphanResourceForCleanup determines if an unannotated resource should be cleaned up as orphan
func (c *K8sClient) isOrphanResourceForCleanup(ctx context.Context, obj client.Object, cleanupConfig *cronschedulesv1.CleanupConfig) bool {
	logger := log.FromContext(ctx)

	// Parse the max age duration
	maxAge, err := time.ParseDuration(cleanupConfig.OrphanResourceMaxAge)
	if err != nil {
		logger.Error(err, "Invalid orphan resource max age format", "maxAge", cleanupConfig.OrphanResourceMaxAge)
		return false
	}

	// Calculate if resource is old enough to be considered orphan
	now := time.Now()
	resourceAge := now.Sub(obj.GetCreationTimestamp().Time)

	if resourceAge > maxAge {
		logger.Info("Orphan resource cleanup time reached",
			"name", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"age", resourceAge,
			"maxAge", maxAge,
			"created", obj.GetCreationTimestamp().Time)
		return true
	}

	logger.V(1).Info("Orphan resource not old enough for cleanup",
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
		"age", resourceAge,
		"maxAge", maxAge)

	return false
}

// isCleanupTimeReached checks if the cleanup time has been reached
func (c *K8sClient) isCleanupTimeReached(ctx context.Context, cleanupValue string, obj client.Object) bool {
	logger := log.FromContext(ctx)
	now := time.Now()

	// Try to parse as duration (e.g., "24h", "7d")
	if duration, err := time.ParseDuration(cleanupValue); err == nil {
		// Use creation time + duration
		cleanupTime := obj.GetCreationTimestamp().Add(duration)
		if now.After(cleanupTime) {
			logger.Info("Resource cleanup time reached (duration-based)",
				"name", obj.GetName(),
				"created", obj.GetCreationTimestamp().Time,
				"duration", cleanupValue,
				"cleanupTime", cleanupTime)
			return true
		}
		return false
	}

	// Try to parse as absolute time (RFC3339)
	if cleanupTime, err := time.Parse(time.RFC3339, cleanupValue); err == nil {
		if now.After(cleanupTime) {
			logger.Info("Resource cleanup time reached (absolute time)",
				"name", obj.GetName(),
				"cleanupTime", cleanupTime)
			return true
		}
		return false
	}

	// Try to parse as simple date format
	if cleanupTime, err := time.Parse("2006-01-02", cleanupValue); err == nil {
		if now.After(cleanupTime) {
			logger.Info("Resource cleanup time reached (date-based)",
				"name", obj.GetName(),
				"cleanupTime", cleanupTime)
			return true
		}
		return false
	}

	logger.Error(nil, "Invalid cleanup time format",
		"name", obj.GetName(),
		"value", cleanupValue,
		"supportedFormats", "duration (24h, 7d), RFC3339 (2006-01-02T15:04:05Z07:00), or date (2006-01-02)")

	return false
}
