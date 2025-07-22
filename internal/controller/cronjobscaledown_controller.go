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
	"regexp"
	"strings"
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

var (
	// Valid timezone regex pattern - restricts to safe IANA timezone names
	validTimezonePattern = regexp.MustCompile(`^[A-Za-z]+(?:[_/][A-Za-z0-9_+-]+)*$`)
	// Maximum schedule length to prevent extremely long schedules
	maxScheduleLength = 100
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
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	return r.processSchedules(ctx, cronJobScaleDown)
}

func (r *CronJobScaleDownReconciler) validateSpec(cronJobScaleDown *cronschedulesv1.CronJobScaleDown) error {
	if cronJobScaleDown.Spec.ScaleDownSchedule == "" &&
		cronJobScaleDown.Spec.ScaleUpSchedule == "" &&
		cronJobScaleDown.Spec.CleanupSchedule == "" {
		return fmt.Errorf("all schedules (ScaleDownSchedule, ScaleUpSchedule, CleanupSchedule) are empty")
	}

	// Validate schedule lengths
	if len(cronJobScaleDown.Spec.ScaleDownSchedule) > maxScheduleLength {
		return fmt.Errorf("ScaleDownSchedule exceeds maximum length of %d characters", maxScheduleLength)
	}
	if len(cronJobScaleDown.Spec.ScaleUpSchedule) > maxScheduleLength {
		return fmt.Errorf("ScaleUpSchedule exceeds maximum length of %d characters", maxScheduleLength)
	}
	if len(cronJobScaleDown.Spec.CleanupSchedule) > maxScheduleLength {
		return fmt.Errorf("CleanupSchedule exceeds maximum length of %d characters", maxScheduleLength)
	}

	// Validate schedule format
	if cronJobScaleDown.Spec.ScaleDownSchedule != "" {
		if err := r.validateCronSchedule(cronJobScaleDown.Spec.ScaleDownSchedule); err != nil {
			return fmt.Errorf("invalid ScaleDownSchedule: %w", err)
		}
	}
	if cronJobScaleDown.Spec.ScaleUpSchedule != "" {
		if err := r.validateCronSchedule(cronJobScaleDown.Spec.ScaleUpSchedule); err != nil {
			return fmt.Errorf("invalid ScaleUpSchedule: %w", err)
		}
	}
	if cronJobScaleDown.Spec.CleanupSchedule != "" {
		if err := r.validateCronSchedule(cronJobScaleDown.Spec.CleanupSchedule); err != nil {
			return fmt.Errorf("invalid CleanupSchedule: %w", err)
		}
	}

	// Validate cleanup configuration if cleanup schedule is provided
	if cronJobScaleDown.Spec.CleanupSchedule != "" {
		if err := r.validateCleanupConfig(cronJobScaleDown.Spec.CleanupConfig); err != nil {
			return fmt.Errorf("invalid CleanupConfig: %w", err)
		}
	}

	// Validate timezone
	if cronJobScaleDown.Spec.TimeZone == "" {
		return fmt.Errorf("TimeZone is empty")
	}
	if err := r.validateTimezone(cronJobScaleDown.Spec.TimeZone); err != nil {
		return fmt.Errorf("invalid TimeZone: %w", err)
	}

	// Validate target reference only if scaling schedules are provided
	if cronJobScaleDown.Spec.ScaleDownSchedule != "" || cronJobScaleDown.Spec.ScaleUpSchedule != "" {
		if cronJobScaleDown.Spec.TargetRef == nil {
			return fmt.Errorf("targetRef is required when scaling schedules are provided")
		}
		if err := r.validateTargetRef(cronJobScaleDown.Spec.TargetRef); err != nil {
			return fmt.Errorf("invalid TargetRef: %w", err)
		}
	}

	return nil
}

func (r *CronJobScaleDownReconciler) validateCronSchedule(schedule string) error {
	// Sanitize schedule - remove potentially dangerous characters
	schedule = strings.TrimSpace(schedule)
	if schedule == "" {
		return fmt.Errorf("schedule cannot be empty after trimming")
	}

	// Check for potentially dangerous patterns
	if strings.Contains(schedule, "..") || strings.Contains(schedule, "//") {
		return fmt.Errorf("schedule contains potentially dangerous patterns")
	}

	// Validate using cron parser
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(schedule)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	return nil
}

func (r *CronJobScaleDownReconciler) validateTimezone(timezone string) error {
	// Sanitize timezone
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return fmt.Errorf("timezone cannot be empty after trimming")
	}

	// Validate timezone format using regex
	if !validTimezonePattern.MatchString(timezone) {
		return fmt.Errorf("timezone contains invalid characters")
	}

	// Test if timezone can be loaded
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	return nil
}

func (r *CronJobScaleDownReconciler) validateTargetRef(targetRef *cronschedulesv1.TargetRef) error {
	if targetRef.Name == "" {
		return fmt.Errorf("target name cannot be empty")
	}
	if targetRef.Namespace == "" {
		return fmt.Errorf("target namespace cannot be empty")
	}
	if targetRef.Kind == "" {
		return fmt.Errorf("target kind cannot be empty")
	}
	if targetRef.ApiVersion == "" {
		return fmt.Errorf("target apiVersion cannot be empty")
	}

	// Validate supported resource kinds
	switch targetRef.Kind {
	case utils.DeploymentKind, utils.StatefulSetKind:
		// Valid kinds
	default:
		return fmt.Errorf("unsupported target kind: %s", targetRef.Kind)
	}

	// Validate API version
	if targetRef.ApiVersion != "apps/v1" {
		return fmt.Errorf("unsupported API version: %s", targetRef.ApiVersion)
	}

	return nil
}

func (r *CronJobScaleDownReconciler) validateCleanupConfig(cleanupConfig *cronschedulesv1.CleanupConfig) error {
	if cleanupConfig == nil {
		return fmt.Errorf("cleanup config is required when cleanup schedule is provided")
	}

	if cleanupConfig.AnnotationKey == "" {
		return fmt.Errorf("cleanup annotation key cannot be empty")
	}

	if len(cleanupConfig.ResourceTypes) == 0 {
		return fmt.Errorf("at least one resource type must be specified for cleanup")
	}

	// Validate supported resource types
	supportedTypes := map[string]bool{
		"Deployment":  true,
		"StatefulSet": true,
		"Service":     true,
		"ConfigMap":   true,
		"Secret":      true,
	}

	for _, resourceType := range cleanupConfig.ResourceTypes {
		if !supportedTypes[resourceType] {
			return fmt.Errorf("unsupported resource type for cleanup: %s", resourceType)
		}
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

	cleanupNext, err := r.parseSchedule(cronJobScaleDown.Spec.CleanupSchedule, now)
	if err != nil {
		logger.Error(err, "Error parsing cleanup schedule", "schedule", cronJobScaleDown.Spec.CleanupSchedule)
		return ctrl.Result{}, nil
	}

	didScale, err := r.executeScaling(ctx, k8sClient, cronJobScaleDown, now, scaleDownNext, scaleUpNext)
	if err != nil {
		return ctrl.Result{}, err
	}

	didCleanup, err := r.executeCleanup(ctx, k8sClient, cronJobScaleDown, now)
	if err != nil {
		logger.Error(err, "Error executing cleanup")
		// Don't return error, just log it and continue
	}

	if didScale || didCleanup {
		if err := r.Status().Update(ctx, cronJobScaleDown); err != nil {
			logger.Error(err, "Error updating CronJobScaleDown status")
			return ctrl.Result{}, err
		}
	}

	return r.calculateRequeue(logger, now, scaleDownNext, scaleUpNext, cleanupNext), nil
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

	// Use the cron parser to determine the next execution time for the schedule
	nextTime := cronSchedule.Next(lastExecutionTime)

	// Check if the current time matches or exceeds the next execution time
	return now.After(nextTime) || now.Equal(nextTime)
}

func (r *CronJobScaleDownReconciler) executeScaling(ctx context.Context, k8sClient *utils.K8sClient, cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now time.Time, scaleDownNext, scaleUpNext time.Time) (bool, error) {
	logger := log.FromContext(ctx)
	var didScale bool

	// Skip scaling if no targetRef is provided
	if cronJobScaleDown.Spec.TargetRef == nil {
		return false, nil
	}

	// Debug logging
	logger.Info("Checking scaling conditions",
		"now", now.Format(time.RFC3339),
		"scaleDownNext", scaleDownNext.Format(time.RFC3339),
		"scaleUpNext", scaleUpNext.Format(time.RFC3339),
		"lastScaleDownTime", cronJobScaleDown.Status.LastScaleDownTime.Time.Format(time.RFC3339),
		"lastScaleUpTime", cronJobScaleDown.Status.LastScaleUpTime.Time.Format(time.RFC3339))

	if r.shouldScaleDown(cronJobScaleDown, now) {
		logger.Info("Scaling down the target resource")
		if err := k8sClient.ScaleDownTargetResource(ctx, utils.TargetObject{TargetRef: *cronJobScaleDown.Spec.TargetRef}); err != nil {
			logger.Error(err, "Error scaling down target resource")
			return false, err
		}
		cronJobScaleDown.Status.LastScaleDownTime = metav1.Time{Time: now}
		r.updateCurrentReplicas(ctx, k8sClient, cronJobScaleDown)
		didScale = true
	}

	if r.shouldScaleUp(cronJobScaleDown, now) {
		logger.Info("Scaling up the target resource")
		if err := k8sClient.ScaleUpTargetResource(ctx, utils.TargetObject{TargetRef: *cronJobScaleDown.Spec.TargetRef}); err != nil {
			logger.Error(err, "Error scaling up target resource")
			return false, err
		}
		cronJobScaleDown.Status.LastScaleUpTime = metav1.Time{Time: now}
		r.updateCurrentReplicas(ctx, k8sClient, cronJobScaleDown)
		didScale = true
	}

	return didScale, nil
}

func (r *CronJobScaleDownReconciler) shouldScaleDown(cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now time.Time) bool {
	return r.shouldExecuteNow(
		cronJobScaleDown.Spec.ScaleDownSchedule,
		now,
		cronJobScaleDown.Status.LastScaleDownTime.Time,
	)
}

func (r *CronJobScaleDownReconciler) shouldScaleUp(cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now time.Time) bool {
	return r.shouldExecuteNow(
		cronJobScaleDown.Spec.ScaleUpSchedule,
		now,
		cronJobScaleDown.Status.LastScaleUpTime.Time,
	)
}

func (r *CronJobScaleDownReconciler) executeCleanup(ctx context.Context, k8sClient *utils.K8sClient, cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now time.Time) (bool, error) {
	logger := log.FromContext(ctx)

	if !r.shouldCleanup(cronJobScaleDown, now) {
		return false, nil
	}

	if cronJobScaleDown.Spec.CleanupConfig == nil {
		logger.Error(nil, "Cleanup config is nil but cleanup should execute")
		return false, fmt.Errorf("cleanup config is nil")
	}

	logger.Info("Executing resource cleanup")

	// Use the CronJobScaleDown's namespace as default
	defaultNamespace := cronJobScaleDown.Namespace
	cleanedCount, err := k8sClient.CleanupResources(ctx, cronJobScaleDown.Spec.CleanupConfig, defaultNamespace)
	if err != nil {
		logger.Error(err, "Error during resource cleanup")
		return false, err
	}

	cronJobScaleDown.Status.LastCleanupTime = metav1.Time{Time: now}
	cronJobScaleDown.Status.LastCleanupResourceCount = cleanedCount

	logger.Info("Cleanup completed", "resourcesCleaned", cleanedCount)
	return true, nil
}

func (r *CronJobScaleDownReconciler) shouldCleanup(cronJobScaleDown *cronschedulesv1.CronJobScaleDown, now time.Time) bool {
	return r.shouldExecuteNow(
		cronJobScaleDown.Spec.CleanupSchedule,
		now,
		cronJobScaleDown.Status.LastCleanupTime.Time,
	)
}

func (r *CronJobScaleDownReconciler) updateCurrentReplicas(ctx context.Context, k8sClient *utils.K8sClient, cronJobScaleDown *cronschedulesv1.CronJobScaleDown) {
	if cronJobScaleDown.Spec.TargetRef != nil {
		if current := k8sClient.GetReplicasCount(ctx, utils.TargetObject{TargetRef: *cronJobScaleDown.Spec.TargetRef}); current != nil {
			cronJobScaleDown.Status.CurrentReplicas = *current
		}
	}
}

func (r *CronJobScaleDownReconciler) calculateRequeue(logger logr.Logger, now time.Time, scaleDownNext, scaleUpNext, cleanupNext time.Time) ctrl.Result {
	nexts := []time.Time{}
	if !scaleDownNext.IsZero() {
		nexts = append(nexts, scaleDownNext)
	}
	if !scaleUpNext.IsZero() {
		nexts = append(nexts, scaleUpNext)
	}
	if !cleanupNext.IsZero() {
		nexts = append(nexts, cleanupNext)
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
