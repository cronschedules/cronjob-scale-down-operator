package utils

import (
	"context"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
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
		o.Spec.Replicas = ptr.To(int32(originalReplicas))
		if err := c.Update(ctx, o); err != nil {
			logger.Error(err, "Failed to scale up deployment", "name", o.GetName())
			return err
		}
		logger.Info("Scaled up deployment", "name", o.GetName(), "replicas", originalReplicas)
	case *appsv1.StatefulSet:
		o.Spec.Replicas = ptr.To(int32(originalReplicas))
		if err := c.Update(ctx, o); err != nil {
			logger.Error(err, "Failed to scale up statefulset", "name", o.GetName())
			return err
		}
		logger.Info("Scaled up statefulset", "name", o.GetName(), "replicas", originalReplicas)
	}

	return nil
}
