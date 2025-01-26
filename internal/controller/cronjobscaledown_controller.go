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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

// CronJobScaleDownReconciler reconciles a CronJobScaleDown object
type CronJobScaleDownReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cronschedules.elbazi.co,resources=cronjobscaledowns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cronschedules.elbazi.co,resources=cronjobscaledowns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cronschedules.elbazi.co,resources=cronjobscaledowns/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *CronJobScaleDownReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// The controller logic :
	// 1. Get the CronJobScaleDown resource
	// 2. Parse the cron schedule and get the next execution time or what is the next time to scale down
	// 3. If the next execution time is in the past, scale down the target resource
	// 4. Update the CronJobScaleDown resource status with the last scale down time
	// 5. If the next execution time is in the future, return and wait for the next execution time

	logger := log.FromContext(ctx)
	logger.Info("Reconciling CronJobScaleDown", "name", req.NamespacedName)

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
	if err != nil {
		return ctrl.Result{}, err
	}

	// Get the next execution time
	nextExecutionTime := schedule.Next(time.Now())

	// If the next execution time is in the past, scale down the target resource
	if nextExecutionTime.Before(time.Now()) {
		// Scale down the target resource
		// Update the CronJobScaleDown resource status with the last scale down time
	}

	// If the next execution time is in the future, return and wait for the next execution time
	//return ctrl.Result{RequeueAfter: nextExecutionTime.Sub(time.Now())}, nil
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobScaleDownReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cronschedulesv1.CronJobScaleDown{}).
		Named("cronjobscaledown").
		Complete(r)
}
