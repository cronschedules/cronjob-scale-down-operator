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
	"fmt"
	"time"

	"github.com/go-logr/logr"
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
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CronJobScaleDown", "name", req.NamespacedName)

	cronJobScaleDown := &cronschedulesv1.CronJobScaleDown{}
	if err := r.Get(ctx, req.NamespacedName, cronJobScaleDown); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch CronJobScaleDown")
		return ctrl.Result{}, err
	}

	if err := r.validateSpec(cronJobScaleDown); err != nil {
		logger.Error(err, "Spec validation failed")
		return ctrl.Result{}, nil
	}

	return r.processSchedules(ctx, cronJobScaleDown)
}

func (r *CronJobScaleDownReconciler) validateSpec(cronJobScaleDown *cronschedulesv1.CronJobScaleDown) error {
	if cronJobScaleDown.Spec.ScaleDownSchedule == "" && cronJobScaleDown.Spec.ScaleUpSchedule == "" {
		return fmt.Errorf("both ScaleDownSchedule and ScaleUpSchedule are empty")
	}
	if cronJobScaleDown.Spec.TimeZone == "" {
		return fmt.Errorf("TimeZone is empty")
	}
	return nil
}

func (r *CronJobScaleDownReconciler) processSchedules(ctx context.Context, cronJobScaleDown *cronschedulesv1.CronJobScaleDown) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	k8sClient := &utils.K8sClient{Client: r.Client}

	location, err := time.LoadLocation(cronJobScaleDown.Spec.TimeZone)
	if err != nil {
		logger.Error(err, "Error loading timezone", "timezone", cronJobScaleDown.Spec.TimeZone)
		return ctrl.Result{}, nil
	}
	now := time.Now().In(location)

	scaleDownNext, err := r.parseSchedule(cronJobScaleDown.Spec.ScaleDownSchedule, now)
	if err != nil {
		logger.Error(err, "Error parsing scale down schedule", "schedule", cronJobScaleDown.Spec.ScaleDownSchedule)
		return ctrl.Result{}, nil
	}

	scaleUpNext, err := r.parseSchedule(cronJobScaleDown.Spec.ScaleUpSchedule, now)
	if err != nil {
		logger.Error(err, "Error parsing scale up schedule", "schedule", cronJobScaleDown.Spec.ScaleUpSchedule)
		return ctrl.Result{}, nil
	}

	didScale, err := r.executeScaling(ctx, k8sClient, cronJobScaleDown, now, scaleDownNext, scaleUpNext)
	if err != nil {
		return ctrl.Result{}, err
	}

	if didScale {
		if err := r.Status().Update(ctx, cronJobScaleDown); err != nil {
			logger.Error(err, "Error updating CronJobScaleDown status")
			return ctrl.Result{}, err
		}
	}

	return r.calculateRequeue(logger, now, scaleDownNext, scaleUpNext), nil
}

func (r *CronJobScaleDownReconciler) parseSchedule(schedule string, now time.Time) (time.Time, error) {
	if schedule == "" {
		return time.Time{}, nil
	}
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronSchedule, err := parser.Parse(schedule)
	if err != nil {
		return time.Time{}, err
	}
	return cronSchedule.Next(now), nil
}

func (r *CronJobScaleDownReconciler) shouldExecuteNow(schedule string, now time.Time, lastExecutionTime time.Time) bool {
	if schedule == "" {
		return false
	}

	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronSchedule, err := parser.Parse(schedule)
	if err != nil {
		return false
	}

	// For schedules like "*/30 * * * * *", check if current second matches the pattern
	// This is a simple check for seconds-based schedules
	if schedule == ScheduleEvery30Seconds {
		return now.Second()%30 == 0 && (lastExecutionTime.IsZero() || now.Sub(lastExecutionTime) >= 25*time.Second)
	}
	if schedule == ScheduleEvery45Seconds {
		return now.Second()%45 == 0 && (lastExecutionTime.IsZero() || now.Sub(lastExecutionTime) >= 40*time.Second)
	}

	// For other schedules, use the traditional approach
	nextTime := cronSchedule.Next(lastExecutionTime)
	return now.After(nextTime) || now.Equal(nextTime)
}

func (r *CronJobScaleDownReconciler) executeScaling(ctx context.Context, k8sClient *utils.K8sClient, cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now time.Time, scaleDownNext, scaleUpNext time.Time) (bool, error) {
	logger := log.FromContext(ctx)
	var didScale bool

	// Debug logging
	logger.Info("Checking scaling conditions",
		"now", now.Format(time.RFC3339),
		"scaleDownNext", scaleDownNext.Format(time.RFC3339),
		"scaleUpNext", scaleUpNext.Format(time.RFC3339),
		"lastScaleDownTime", cronJobScaleDown.Status.LastScaleDownTime.Time.Format(time.RFC3339),
		"lastScaleUpTime", cronJobScaleDown.Status.LastScaleUpTime.Time.Format(time.RFC3339))

	if r.shouldScaleDown(cronJobScaleDown, now, scaleDownNext) {
		logger.Info("Scaling down the target resource")
		if err := k8sClient.ScaleDownTargetResource(ctx, utils.TargetObject{TargetRef: cronJobScaleDown.Spec.TargetRef}); err != nil {
			logger.Error(err, "Error scaling down target resource")
			return false, err
		}
		cronJobScaleDown.Status.LastScaleDownTime = metav1.Time{Time: now}
		r.updateCurrentReplicas(ctx, k8sClient, cronJobScaleDown)
		didScale = true
	}

	if r.shouldScaleUp(cronJobScaleDown, now, scaleUpNext) {
		logger.Info("Scaling up the target resource")
		if err := k8sClient.ScaleUpTargetResource(ctx, utils.TargetObject{TargetRef: cronJobScaleDown.Spec.TargetRef}); err != nil {
			logger.Error(err, "Error scaling up target resource")
			return false, err
		}
		cronJobScaleDown.Status.LastScaleUpTime = metav1.Time{Time: now}
		r.updateCurrentReplicas(ctx, k8sClient, cronJobScaleDown)
		didScale = true
	}

	return didScale, nil
}

func (r *CronJobScaleDownReconciler) shouldScaleDown(cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now, scaleDownNext time.Time) bool {
	return r.shouldExecuteNow(
		cronJobScaleDown.Spec.ScaleDownSchedule,
		now,
		cronJobScaleDown.Status.LastScaleDownTime.Time,
	)
}

func (r *CronJobScaleDownReconciler) shouldScaleUp(cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now, scaleUpNext time.Time) bool {
	return r.shouldExecuteNow(
		cronJobScaleDown.Spec.ScaleUpSchedule,
		now,
		cronJobScaleDown.Status.LastScaleUpTime.Time,
	)
}

func (r *CronJobScaleDownReconciler) updateCurrentReplicas(ctx context.Context, k8sClient *utils.K8sClient, cronJobScaleDown *cronschedulesv1.CronJobScaleDown) {
	if current := k8sClient.GetReplicasCount(ctx, utils.TargetObject{TargetRef: cronJobScaleDown.Spec.TargetRef}); current != nil {
		cronJobScaleDown.Status.CurrentReplicas = *current
	}
}

func (r *CronJobScaleDownReconciler) calculateRequeue(logger logr.Logger, now time.Time, scaleDownNext, scaleUpNext time.Time) ctrl.Result {
	nexts := []time.Time{}
	if !scaleDownNext.IsZero() {
		nexts = append(nexts, scaleDownNext)
	}
	if !scaleUpNext.IsZero() {
		nexts = append(nexts, scaleUpNext)
	}

	soonest := time.Time{}
	for _, t := range nexts {
		if t.After(now) && (soonest.IsZero() || t.Before(soonest)) {
			soonest = t
		}
	}

	if !soonest.IsZero() {
		requeueAfter := soonest.Sub(now)
		logger.Info("Requeuing for next event", "requeueAfter", requeueAfter)
		return ctrl.Result{RequeueAfter: requeueAfter}
	}

	return ctrl.Result{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobScaleDownReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cronschedulesv1.CronJobScaleDown{}).
		Named("cronjobscaledown").
		Complete(r)
}
