package utils

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

// K8sClient wraps a kubernetes client
type K8sClient struct {
	client.Client
}

//Documentation of the logic:
// 1. Get the target resource (deployment or statefulset)
// 2. Scale down the target resource to 0 replicas
// 3. Update the target resource status with the last scale down time
// 4. If the next execution time is in the future, return and wait for the next execution time

func scaleDown(ctx context.Context, targetRef cronschedulesv1.TargetRef) error {

	logger := log.FromContext(ctx)
	logger.Info("Scaling down the target resource", "targetRef", targetRef)

	// Get the target resource
	targetResource, err := getTargetResource(ctx, targetRef)
	if err != nil {
		logger.Error(err, "Error getting target resource")
		return err
	}

	fmt.Println("targetResource", targetResource)

	return nil

}

func scaleUp(ctx context.Context, targetRef cronschedulesv1.TargetRef) error {

	logger := log.FromContext(ctx)
	logger.Info("Scaling up the target resource", "targetRef", targetRef)

	// Get the target resource
	targetResource, err := getTargetResource(ctx, targetRef)
	if err != nil {
		logger.Error(err, "Error getting target resource")
		return err
	}

	// Scale up the target resource
	_, err = scaleUpTargetResource(ctx, targetResource)
	if err != nil {
		logger.Error(err, "Error scaling up target resource")
		return err
	}

	return nil
}

func getTargetResource(ctx context.Context, targetRef cronschedulesv1.TargetRef) (client.Object, error) {

	return nil, nil
}

func (c *K8sClient) scaleDownTargetResource(ctx context.Context, targetResource client.Object) (client.Object, error) {
	logger := log.FromContext(ctx)

	if targetResource.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
		// Get the deployment
		deployment := &appsv1.Deployment{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.GetName(), Namespace: targetResource.GetNamespace()}, deployment)
		if err != nil {
			logger.Error(err, "Error getting deployment from the cluster")
			return nil, err
		}

		// Scale down the deployment
		_, err = scaleDownDeployment(ctx, deployment)
		if err != nil {
			logger.Error(err, "Error scaling down deployment")
			return nil, err
		}
		logger.Info("Deployment %s scaled down successfully", deployment.GetName())

	} else if targetResource.GetObjectKind().GroupVersionKind().Kind == "StatefulSet" {
		// Get the statefulset
		statefulset := &appsv1.StatefulSet{}
		err := c.Get(ctx, client.ObjectKey{Name: targetResource.GetName(), Namespace: targetResource.GetNamespace()}, statefulset)
		if err != nil {
			logger.Error(err, "Error getting statefulset from the cluster")
			return nil, err
		}

		// Scale down the statefulset
		_, err = scaleDownStatefulset(ctx, statefulset)
		if err != nil {
			logger.Error(err, "Error scaling down statefulset")
			return nil, err
		}
		logger.Info("Statefulset %s scaled down successfully", statefulset.GetName())
	} else {
		logger.Error(fmt.Errorf("unsupported target resource kind: %s", targetResource.GetObjectKind().GroupVersionKind().Kind), "Unsupported target resource kind")
		return nil, fmt.Errorf("unsupported target resource kind: %s", targetResource.GetObjectKind().GroupVersionKind().Kind)
	}

	return nil, nil
}

func scaleUpTargetResource(ctx context.Context, targetResource client.Object) (client.Object, error) {

	return nil, nil
}

func (c *K8sClient) retrieveClusterDeployment(ctx context.Context, targetResource client.Object) (client.Object, error) {

	logger := log.FromContext(ctx)

	// Get the deployment
	deployment := &appsv1.Deployment{}
	err := c.Get(ctx, client.ObjectKey{Name: targetResource.GetName(), Namespace: targetResource.GetNamespace()}, deployment)
	if err != nil {
		logger.Error(err, "Error getting deployment")
		return nil, err
	}

	// Check if the deployment is valid
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
		logger.Info("Deployment is already scaled down, skipping scale down")
		return nil, nil
	}

	return deployment, nil
}

func scaleDownDeployment(ctx context.Context, deployment client.Object) (client.Object, error) {
	// TODO: Implement scale down logic
	return deployment, nil
}

func scaleDownStatefulset(ctx context.Context, statefulset client.Object) (client.Object, error) {
	// TODO: Implement scale down logic for statefulset
	return statefulset, nil
}
