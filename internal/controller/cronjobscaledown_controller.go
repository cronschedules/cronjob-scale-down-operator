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

package controller

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
	"github.com/z4ck404/cronjob-scale-down-operator/internal/utils"
)

// CronJobScaleDownReconciler reconciles a CronJobScaleDown object
type CronJobScaleDownReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CronJobScaleDownReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// The controller logic :
	// 1. Get the CronJobScaleDown resource
	// 2. Parse the cron schedule and get the next execution time or what is the next time to scale down
	// 3. If the next execution time is in the past, scale down the target resource
	// 4. Update the CronJobScaleDown resource status with the last scale down time
	// 5. If the next execution time is in the future, return and wait for the next execution time

	logger := log.FromContext(ctx)
	logger.Info("Reconciling CronJobScaleDown", "name", req.NamespacedName)
	k8sClient := &utils.K8sClient{Client: r.Client}

	// Get the CronJobScaleDown resource
	cronJobScaleDown := &cronschedulesv1.CronJobScaleDown{}
	if err := r.Get(ctx, req.NamespacedName, cronJobScaleDown); err != nil {
		return ctrl.Result{}, err
	}

	// Parse the cron schedule and get the next execution time or what is the next time to scale down
	// Create a new cron parser
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	// Parse the cron schedule
	schedule, err := parser.Parse(cronJobScaleDown.Spec.ScaleDownSchedule)
	logger.Info("Schedule", schedule)
	if err != nil {
		logger.Error(err, "Error parsing schedule")
		return ctrl.Result{}, err
	}
	// Get the timezone
	location, err := time.LoadLocation(cronJobScaleDown.Spec.TimeZone)
	logger.Info("Timezone", "timezone", location)
	if err != nil {
		logger.Error(err, "Error loading timezone")
		return ctrl.Result{}, err
	}

	// Get the next execution time
	nextExecutionTime := schedule.Next(time.Now().In(location))
	logger.Info("Next execution time", nextExecutionTime)
	// If the next execution time is in the past, scale down the target resource
	if nextExecutionTime.Before(time.Now().In(location)) {
		logger.Info("Next execution time is in the past, scaling down the target resource")
		err := k8sClient.ScaleDownTargetResource(ctx, utils.TargetObject{TargetRef: cronJobScaleDown.Spec.TargetRef})
		if err != nil {
			logger.Error(err, "Error scaling down target resource")
			return ctrl.Result{}, err
		}
		// Update the CronJobScaleDown resource status with the last scale down time
		cronJobScaleDown.Status.LastScaleDownTime = metav1.Time{Time: time.Now().In(location)}
		err = r.Update(ctx, cronJobScaleDown)
		if err != nil {
			logger.Error(err, "Error updating CronJobScaleDown resource status")
			return ctrl.Result{}, err
		}
	} else {
		err := k8sClient.UpdateTargetResourceOriginalReplicasAnnotation(ctx, utils.TargetObject{TargetRef: cronJobScaleDown.Spec.TargetRef})
		if err != nil {
			logger.Error(err, "Error updating target resource original replicas annotation")
			return ctrl.Result{}, err
		}
		logger.Info("Next execution time is in the future, waiting for the next execution time")
		return ctrl.Result{RequeueAfter: nextExecutionTime.Sub(time.Now().In(location))}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobScaleDownReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cronschedulesv1.CronJobScaleDown{}).
		Named("cronjobscaledown").
		Complete(r)
}
