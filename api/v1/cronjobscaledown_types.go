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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobScaleDownSpec defines the desired state of CronJobScaleDown.
type CronJobScaleDownSpec struct {
	// Target resource to scale (Deployment/StatefulSet)
	// +kubebuilder:validation:Optional
	TargetRef *TargetRef `json:"targetRef,omitempty"`

	// Cron schedule for scaling down (e.g., "0 22 * * *" for 10 PM daily)
	// +kubebuilder:validation:Optional
	ScaleDownSchedule string `json:"scaleDownSchedule,omitempty"`

	// Cron schedule for scaling back up (e.g., "0 6 * * *" for 6 AM daily)
	// +kubebuilder:validation:Optional
	ScaleUpSchedule string `json:"scaleUpSchedule,omitempty"`

	// Cron schedule for cleaning up resources (e.g., "0 0 * * 0" for every Sunday)
	// +kubebuilder:validation:Optional
	CleanupSchedule string `json:"cleanupSchedule,omitempty"`

	// Cleanup configuration for deleting resources based on annotations
	// +kubebuilder:validation:Optional
	CleanupConfig *CleanupConfig `json:"cleanupConfig,omitempty"`

	// Timezone (e.g., "America/New_York", "UTC")
	// +kubebuilder:validation:Required
	// +kubebuilder:default:="UTC"
	TimeZone string `json:"timeZone"`
}

type TargetRef struct {
	// Name of the target resource
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Namespace of the target resource
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	// Kind of the target resource (Deployment, StatefulSet)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Deployment;StatefulSet
	Kind string `json:"kind"`
	// ApiVersion of the target resource
	// +kubebuilder:validation:Required
	// +kubebuilder:default:="apps/v1"
	ApiVersion string `json:"apiVersion"`
}

type CleanupConfig struct {
	// Namespaces to search for resources to cleanup (defaults to same namespace as the CronJobScaleDown)
	// +kubebuilder:validation:Optional
	Namespaces []string `json:"namespaces,omitempty"`

	// Annotation key that marks resources for cleanup
	// +kubebuilder:validation:Required
	AnnotationKey string `json:"annotationKey"`

	// Resource types to cleanup (e.g., ["Deployment", "StatefulSet", "Service", "ConfigMap"])
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	ResourceTypes []string `json:"resourceTypes"`

	// Label selector to further filter resources for cleanup
	// +kubebuilder:validation:Optional
	LabelSelector map[string]string `json:"labelSelector,omitempty"`

	// DryRun mode - if true, only logs what would be deleted without actually deleting
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	DryRun bool `json:"dryRun,omitempty"`
}

// CronJobScaleDownStatus defines the observed state of CronJobScaleDown.
type CronJobScaleDownStatus struct {
	// LastScaleDownTime is the time when the scale down was last performed
	LastScaleDownTime metav1.Time `json:"lastScaleDownTime,omitempty"`

	// LastScaleUpTime is the time when the scale up was last performed
	LastScaleUpTime metav1.Time `json:"lastScaleUpTime,omitempty"`

	// LastCleanupTime is the time when the cleanup was last performed
	LastCleanupTime metav1.Time `json:"lastCleanupTime,omitempty"`

	// CurrentReplicas is the current number of replicas
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	// LastCleanupResourceCount is the number of resources cleaned up in the last cleanup operation
	LastCleanupResourceCount int32 `json:"lastCleanupResourceCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CronJobScaleDown is the Schema for the cronjobscaledowns API.
type CronJobScaleDown struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobScaleDownSpec   `json:"spec,omitempty"`
	Status CronJobScaleDownStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronJobScaleDownList contains a list of CronJobScaleDown.
type CronJobScaleDownList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJobScaleDown `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJobScaleDown{}, &CronJobScaleDownList{})
}
