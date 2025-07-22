package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

// K8sClient wraps a kubernetes client
type K8sClient struct {
	client.Client
}

type TargetObject struct {
	cronschedulesv1.TargetRef
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
	deployment.Spec.Replicas = ptr.To[int32](0)
	return c.Update(ctx, deployment)
}

func (c *K8sClient) scaleDownStatefulset(ctx context.Context, statefulset *appsv1.StatefulSet) error {
	statefulset.Spec.Replicas = ptr.To[int32](0)
	return c.Update(ctx, statefulset)
}

func (c *K8sClient) GetReplicasCount(ctx context.Context, targetResource TargetObject) *int32 {
	logger := log.FromContext(ctx)
	var replicas *int32

	switch targetResource.Kind {
	case DeploymentKind:
		deployment := &appsv1.Deployment{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, deployment)
		if err != nil {
			logger.Error(err, "Error getting deployment from the cluster", "name", targetResource.Name)
			return nil
		}
		replicas = deployment.Spec.Replicas

	case StatefulSetKind:
		statefulset := &appsv1.StatefulSet{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, statefulset)
		if err != nil {
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

	if err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, targetResourceObject); err != nil {
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
		logger.Error(err, "Failed to update target resource original replicas annotation", "name", targetResource.Name)
		return fmt.Errorf("failed to update target resource original replicas annotation: %w", err)
	}

	logger.Info("Set original replicas annotation", "name", targetResource.Name, "replicas", *originalTargetResourceReplicas)
	return nil
}

// ScaleUpTargetResource scales up the target resource to its original replica count (from annotation)
func (c *K8sClient) ScaleUpTargetResource(ctx context.Context, targetRef TargetObject) error {
	logger := log.FromContext(ctx)

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

	if err := c.Get(ctx, client.ObjectKey{Name: targetRef.Name, Namespace: targetRef.Namespace}, obj); err != nil {
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
			logger.Error(err, "Failed to scale up deployment", "name", o.GetName())
			return err
		}
		logger.Info("Successfully scaled up deployment", "name", o.GetName(), "replicas", originalReplicas)
	case *appsv1.StatefulSet:
		o.Spec.Replicas = ptr.To[int32](int32(originalReplicas))
		if err := c.Update(ctx, o); err != nil {
			logger.Error(err, "Failed to scale up statefulset", "name", o.GetName())
			return err
		}
		logger.Info("Successfully scaled up statefulset", "name", o.GetName(), "replicas", originalReplicas)
	default:
		logger.Error(nil, "Unsupported resource type for scaling", "type", fmt.Sprintf("%T", obj))
		return fmt.Errorf("unsupported resource type: %T", obj)
	}

	return nil
}

// CleanupResources finds and deletes resources based on cleanup configuration
func (c *K8sClient) CleanupResources(ctx context.Context, cleanupConfig *cronschedulesv1.CleanupConfig, defaultNamespace string) (int32, error) {
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
			deleted, err := c.cleanupResourceType(ctx, resourceType, namespace, cleanupConfig)
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

// cleanupResourceType handles cleanup for a specific resource type in a namespace
func (c *K8sClient) cleanupResourceType(ctx context.Context, resourceType, namespace string, cleanupConfig *cronschedulesv1.CleanupConfig) (int32, error) {
	var objList client.ObjectList

	// Map resource types to their corresponding types
	switch resourceType {
	case "Deployment":
		objList = &appsv1.DeploymentList{}
	case "StatefulSet":
		objList = &appsv1.StatefulSetList{}
	case "Service":
		objList = &corev1.ServiceList{}
	case "ConfigMap":
		objList = &corev1.ConfigMapList{}
	case "Secret":
		objList = &corev1.SecretList{}
	default:
		return 0, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Prepare list options
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
	}

	// Add label selector if specified
	if len(cleanupConfig.LabelSelector) > 0 {
		selector := labels.SelectorFromSet(cleanupConfig.LabelSelector)
		listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: selector})
	}

	// List resources
	if err := c.List(ctx, objList, listOpts...); err != nil {
		return 0, fmt.Errorf("failed to list %s in namespace %s: %w", resourceType, namespace, err)
	}

	var deleted int32

	// Process each resource based on type
	switch list := objList.(type) {
	case *appsv1.DeploymentList:
		for _, item := range list.Items {
			if c.shouldCleanupResource(ctx, &item, cleanupConfig) {
				deleted += c.deleteResource(ctx, &item, cleanupConfig.DryRun)
			}
		}
	case *appsv1.StatefulSetList:
		for _, item := range list.Items {
			if c.shouldCleanupResource(ctx, &item, cleanupConfig) {
				deleted += c.deleteResource(ctx, &item, cleanupConfig.DryRun)
			}
		}
	case *corev1.ServiceList:
		for _, item := range list.Items {
			if c.shouldCleanupResource(ctx, &item, cleanupConfig) {
				deleted += c.deleteResource(ctx, &item, cleanupConfig.DryRun)
			}
		}
	case *corev1.ConfigMapList:
		for _, item := range list.Items {
			if c.shouldCleanupResource(ctx, &item, cleanupConfig) {
				deleted += c.deleteResource(ctx, &item, cleanupConfig.DryRun)
			}
		}
	case *corev1.SecretList:
		for _, item := range list.Items {
			if c.shouldCleanupResource(ctx, &item, cleanupConfig) {
				deleted += c.deleteResource(ctx, &item, cleanupConfig.DryRun)
			}
		}
	}

	return deleted, nil
}

// deleteResource handles the actual deletion or dry-run logging
func (c *K8sClient) deleteResource(ctx context.Context, obj client.Object, dryRun bool) int32 {
	logger := log.FromContext(ctx)

	// Get resource type from the object type
	resourceType := fmt.Sprintf("%T", obj)

	if dryRun {
		logger.Info("DRY RUN: Would delete resource",
			"type", resourceType,
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())
		return 1
	} else {
		if err := c.Delete(ctx, obj); err != nil {
			logger.Error(err, "Failed to delete resource",
				"type", resourceType,
				"name", obj.GetName(),
				"namespace", obj.GetNamespace())
			return 0
		}
		logger.Info("Successfully deleted resource",
			"type", resourceType,
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())
		return 1
	}
}

// shouldCleanupResource determines if a resource should be cleaned up based on annotations
func (c *K8sClient) shouldCleanupResource(ctx context.Context, obj client.Object, cleanupConfig *cronschedulesv1.CleanupConfig) bool {
	logger := log.FromContext(ctx)

	annotations := obj.GetAnnotations()
	if annotations == nil {
		return false
	}

	// Check if cleanup annotation exists
	cleanupValue, exists := annotations[cleanupConfig.AnnotationKey]
	if !exists {
		return false
	}

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
