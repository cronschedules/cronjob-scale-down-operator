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
	case "Deployment":
		deployment := &appsv1.Deployment{}
		err := c.Get(ctx, client.ObjectKey{Name: targetRef.Name, Namespace: targetRef.Namespace}, deployment)
		if err != nil {
			logger.Error(err, "Error getting deployment %s from the cluster", targetRef.Name)
			return err
		}

		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
			logger.Info("Deployment %s is already scaled down, skipping scale down", deployment.GetName())
			return nil
		}

		err = c.scaleDownDeployment(ctx, deployment)
		if err != nil {
			logger.Error(err, "Error scaling down deployment")
			return err
		}

		logger.Info("Deployment %s scaled down successfully", deployment.GetName())

	case "StatefulSet":
		statefulset := &appsv1.StatefulSet{}
		err := c.Get(ctx, client.ObjectKey{Name: targetRef.Name, Namespace: targetRef.Namespace}, statefulset)
		if err != nil {
			logger.Error(err, "Error getting statefulset from the cluster")
			return err
		}

		if statefulset.Spec.Replicas != nil && *statefulset.Spec.Replicas == 0 {
			logger.Info("Statefulset %s is already scaled down, skipping scale down", statefulset.GetName())
			return nil
		}

		err = c.scaleDownStatefulset(ctx, statefulset)
		if err != nil {
			logger.Error(err, "Error scaling down statefulset %s", statefulset.GetName())
			return err
		}

		logger.Info("Statefulset %s scaled down successfully", statefulset.GetName())
	default:
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
	var replicas *int32

	switch targetResource.Kind {
	case "Deployment":
		deployment := &appsv1.Deployment{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, deployment)
		if err != nil {
			fmt.Println("Error getting deployment from the cluster")
			return nil
		}
		replicas = deployment.Spec.Replicas

	case "StatefulSet":
		statefulset := &appsv1.StatefulSet{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, statefulset)
		if err != nil {
			fmt.Println("Error getting statefulset from the cluster")
			return nil
		}
		replicas = statefulset.Spec.Replicas
	default:
		fmt.Println("Unsupported target resource kind")
	}

	return replicas
}

func (c *K8sClient) UpdateTargetResourceOriginalReplicasAnnotation(ctx context.Context, targetResource TargetObject) error {
	var targetResourceObject client.Object

	switch targetResource.Kind {
	case "Deployment":
		targetResourceObject = &appsv1.Deployment{}
	case "StatefulSet":
		targetResourceObject = &appsv1.StatefulSet{}
	default:
		return fmt.Errorf("unsupported target resource kind: %s", targetResource.Kind)
	}

	if err := c.Get(ctx, client.ObjectKey{Name: targetResource.Name, Namespace: targetResource.Namespace}, targetResourceObject); err != nil {
		return fmt.Errorf("failed to get target resource: %w", err)
	}

	annotations := targetResourceObject.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if _, ok := annotations[annotationKeyOriginalReplicas]; ok {
		fmt.Println("Original replicas annotation already exists")
		return nil
	}

	originalTargetResourceReplicas := c.GetReplicasCount(ctx, targetResource)
	if originalTargetResourceReplicas == nil {
		return fmt.Errorf("failed to get original replicas count for target resource")
	}

	annotations[annotationKeyOriginalReplicas] = strconv.Itoa(int(*originalTargetResourceReplicas))
	targetResourceObject.SetAnnotations(annotations)

	if err := c.Update(ctx, targetResourceObject); err != nil {
		return fmt.Errorf("failed to update target resource original replicas annotation: %w", err)
	}

	return nil
}
