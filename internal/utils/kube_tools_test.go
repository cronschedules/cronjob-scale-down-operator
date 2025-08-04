package utils

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

func TestIsOrphanResourceForCleanup(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// Create a fake client
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	k8sClient := &K8sClient{Client: fakeClient}

	// Set up context with logger
	ctx := log.IntoContext(context.Background(), log.Log)

	tests := []struct {
		name          string
		resourceAge   time.Duration
		maxAge        string
		shouldCleanup bool
		description   string
	}{
		{
			name:          "Old resource should be cleaned up",
			resourceAge:   48 * time.Hour, // 2 days old
			maxAge:        "24h",          // max age 1 day
			shouldCleanup: true,
			description:   "Resource older than max age should be cleaned up",
		},
		{
			name:          "New resource should not be cleaned up",
			resourceAge:   12 * time.Hour, // 12 hours old
			maxAge:        "24h",          // max age 1 day
			shouldCleanup: false,
			description:   "Resource younger than max age should not be cleaned up",
		},
		{
			name:          "Resource exactly at max age should be cleaned up",
			resourceAge:   24 * time.Hour, // exactly 1 day old
			maxAge:        "24h",          // max age 1 day
			shouldCleanup: true,
			description:   "Resource exactly at max age should be cleaned up",
		},
		{
			name:          "Very old resource with week max age",
			resourceAge:   10 * 24 * time.Hour, // 10 days old
			maxAge:        "168h",              // max age 1 week (7 days)
			shouldCleanup: true,
			description:   "Very old resource should be cleaned up",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a ConfigMap with the specified age
			now := time.Now()
			creationTime := now.Add(-tt.resourceAge)

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-configmap",
					Namespace:         "default",
					CreationTimestamp: metav1.Time{Time: creationTime},
				},
				Data: map[string]string{
					"key": "value",
				},
			}

			cleanupConfig := &cronschedulesv1.CleanupConfig{
				CleanupOrphanResources: true,
				OrphanResourceMaxAge:   tt.maxAge,
			}

			result := k8sClient.isOrphanResourceForCleanup(ctx, configMap, cleanupConfig)

			if result != tt.shouldCleanup {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.shouldCleanup, result)
			}
		})
	}
}

func TestShouldCleanupResource(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// Create a fake client
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	k8sClient := &K8sClient{Client: fakeClient}

	// Set up context with logger
	ctx := log.IntoContext(context.Background(), log.Log)

	tests := []struct {
		name          string
		annotations   map[string]string
		resourceAge   time.Duration
		orphanEnabled bool
		orphanMaxAge  string
		shouldCleanup bool
		description   string
	}{
		{
			name: "Resource with immediate cleanup annotation",
			annotations: map[string]string{
				"cleanup-after": "",
			},
			shouldCleanup: true,
			description:   "Resource with empty cleanup annotation should be cleaned up immediately",
		},
		{
			name: "Resource without annotation, orphan cleanup disabled",
			annotations: map[string]string{
				"other-annotation": "value",
			},
			orphanEnabled: false,
			shouldCleanup: false,
			description:   "Resource without cleanup annotation should not be cleaned up when orphan cleanup is disabled",
		},
		{
			name:          "Old resource without annotation, orphan cleanup enabled",
			annotations:   map[string]string{},
			resourceAge:   48 * time.Hour, // 2 days old
			orphanEnabled: true,
			orphanMaxAge:  "24h", // max age 1 day
			shouldCleanup: true,
			description:   "Old resource without cleanup annotation should be cleaned up when orphan cleanup is enabled",
		},
		{
			name:          "New resource without annotation, orphan cleanup enabled",
			annotations:   map[string]string{},
			resourceAge:   12 * time.Hour, // 12 hours old
			orphanEnabled: true,
			orphanMaxAge:  "24h", // max age 1 day
			shouldCleanup: false,
			description:   "New resource without cleanup annotation should not be cleaned up even when orphan cleanup is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a ConfigMap with the specified age and annotations
			now := time.Now()
			creationTime := now.Add(-tt.resourceAge)

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-configmap",
					Namespace:         "default",
					CreationTimestamp: metav1.Time{Time: creationTime},
					Annotations:       tt.annotations,
				},
				Data: map[string]string{
					"key": "value",
				},
			}

			cleanupConfig := &cronschedulesv1.CleanupConfig{
				AnnotationKey:          "cleanup-after",
				CleanupOrphanResources: tt.orphanEnabled,
				OrphanResourceMaxAge:   tt.orphanMaxAge,
			}

			result := k8sClient.shouldCleanupResource(ctx, configMap, cleanupConfig)

			if result != tt.shouldCleanup {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.shouldCleanup, result)
			}
		})
	}
}
