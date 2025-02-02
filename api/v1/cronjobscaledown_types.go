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
	TargetRef TargetRef `json:"targetRef"`

	// Cron schedule for scaling down (e.g., "0 22 * * *" for 10 PM daily)
	ScaleDownSchedule string `json:"scaleDownSchedule"`

	// Cron schedule for scaling back up (e.g., "0 6 * * *" for 6 AM daily)
	ScaleUpSchedule string `json:"scaleUpSchedule"`

	// Optional: Timezone (e.g., "America/New_York")
	TimeZone string `json:"timeZone,omitempty"`
}

type TargetRef struct {
	// Name of the target resource
	Name string `json:"name"`
	// Namespace of the target resource
	Namespace string `json:"namespace"`
	// Kind of the target resource (Deployment, StatefulSet)
	Kind string `json:"kind"`
	// ApiVersion of the target resource
	ApiVersion string `json:"apiVersion"`
}

// CronJobScaleDownStatus defines the observed state of CronJobScaleDown.
type CronJobScaleDownStatus struct {
	// LastScaleDownTime is the time when the scale down was last performed
	LastScaleDownTime metav1.Time `json:"lastScaleDownTime,omitempty"`

	// LastScaleUpTime is the time when the scale up was last performed
	LastScaleUpTime metav1.Time `json:"lastScaleUpTime,omitempty"`

	// CurrentReplicas is the current number of replicas
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`
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
