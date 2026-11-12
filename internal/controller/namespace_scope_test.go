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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
	"github.com/z4ck404/cronjob-scale-down-operator/internal/utils"
)

const (
	tenantNamespace = "tenant-a"
	victimNamespace = "victim"
)

func newScopeTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
	Expect(cronschedulesv1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func scalingCR(targetNamespace string) *cronschedulesv1.CronJobScaleDown {
	return &cronschedulesv1.CronJobScaleDown{
		ObjectMeta: metav1.ObjectMeta{Name: "scaler", Namespace: tenantNamespace},
		Spec: cronschedulesv1.CronJobScaleDownSpec{
			TargetRef: &cronschedulesv1.TargetRef{
				Name:       "app",
				Namespace:  targetNamespace,
				Kind:       utils.DeploymentKind,
				ApiVersion: "apps/v1",
			},
			ScaleDownSchedule: "* * * * * *",
			TimeZone:          "UTC",
		},
	}
}

func cleanupCR(namespaces, resourceTypes []string) *cronschedulesv1.CronJobScaleDown {
	return &cronschedulesv1.CronJobScaleDown{
		ObjectMeta: metav1.ObjectMeta{Name: "cleaner", Namespace: tenantNamespace},
		Spec: cronschedulesv1.CronJobScaleDownSpec{
			CleanupSchedule: "* * * * * *",
			CleanupConfig: &cronschedulesv1.CleanupConfig{
				Namespaces:             namespaces,
				AnnotationKey:          "cleanup/after",
				ResourceTypes:          resourceTypes,
				CleanupOrphanResources: true,
				OrphanResourceMaxAge:   "1s",
			},
			TimeZone: "UTC",
		},
	}
}

func deploymentIn(namespace string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "app"}},
		},
	}
}

func oldSecret(namespace string) *corev1.Secret {
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:              "credentials",
		Namespace:         namespace,
		CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Hour)),
	}}
}

var _ = Describe("CronJobScaleDown namespace scope policy", func() {
	Context("validateSpec", func() {
		DescribeTable("enforces namespace and cluster-scope policy",
			func(cr *cronschedulesv1.CronJobScaleDown, allowCrossNamespace, allowClusterScoped, expectErr bool) {
				r := &CronJobScaleDownReconciler{
					AllowCrossNamespace:       allowCrossNamespace,
					AllowClusterScopedCleanup: allowClusterScoped,
				}
				err := r.validateSpec(cr)
				if expectErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("same-namespace target is allowed", scalingCR(tenantNamespace), false, false, false),
			Entry("cross-namespace target is rejected by default", scalingCR(victimNamespace), false, false, true),
			Entry("cross-namespace target is allowed with --allow-cross-namespace", scalingCR(victimNamespace), true, false, false),
			Entry("cleanup without namespaces defaults to own namespace", cleanupCR(nil, []string{"Secret"}), false, false, false),
			Entry("cleanup of own namespace is allowed", cleanupCR([]string{tenantNamespace}, []string{"Secret"}), false, false, false),
			Entry("cleanup of another namespace is rejected by default", cleanupCR([]string{tenantNamespace, victimNamespace}, []string{"Secret"}), false, false, true),
			Entry("cleanup of another namespace is allowed with --allow-cross-namespace", cleanupCR([]string{victimNamespace}, []string{"Secret"}), true, false, false),
			Entry("ClusterRole cleanup is rejected by default", cleanupCR(nil, []string{"ClusterRole"}), false, false, true),
			Entry("ClusterRoleBinding cleanup is rejected even with --allow-cross-namespace", cleanupCR(nil, []string{"ClusterRoleBinding"}), true, false, true),
			Entry("ClusterRole cleanup is allowed with --allow-cluster-scoped-cleanup", cleanupCR(nil, []string{"ClusterRole"}), false, true, false),
		)
	})

	Context("Reconcile with a cross-namespace targetRef", func() {
		run := func(allowCrossNamespace bool) int32 {
			ctx := context.Background()
			cr := scalingCR(victimNamespace)
			fakeClient := fake.NewClientBuilder().
				WithScheme(newScopeTestScheme()).
				WithStatusSubresource(&cronschedulesv1.CronJobScaleDown{}).
				WithObjects(cr, deploymentIn(victimNamespace, 3)).
				Build()
			r := &CronJobScaleDownReconciler{Client: fakeClient, Scheme: fakeClient.Scheme(), AllowCrossNamespace: allowCrossNamespace}

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(cr)})
			Expect(err).NotTo(HaveOccurred())

			victim := &appsv1.Deployment{}
			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: "app", Namespace: victimNamespace}, victim)).To(Succeed())
			return *victim.Spec.Replicas
		}

		It("does not scale the other namespace's Deployment by default", func() {
			Expect(run(false)).To(Equal(int32(3)))
		})

		It("scales it only when --allow-cross-namespace is set", func() {
			Expect(run(true)).To(Equal(int32(0)))
		})
	})

	Context("executeCleanup sink-side enforcement", func() {
		It("only deletes in the CronJobScaleDown namespace even if validation is bypassed", func() {
			ctx := context.Background()
			cr := cleanupCR([]string{victimNamespace, tenantNamespace}, []string{"Secret", "ClusterRole"})
			clusterRole := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{
				Name:              "cluster-admin-like",
				CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Hour)),
			}}
			fakeClient := fake.NewClientBuilder().
				WithScheme(newScopeTestScheme()).
				WithObjects(oldSecret(tenantNamespace), oldSecret(victimNamespace), clusterRole).
				Build()
			r := &CronJobScaleDownReconciler{Client: fakeClient, Scheme: fakeClient.Scheme()}

			_, err := r.executeCleanup(ctx, &utils.K8sClient{Client: fakeClient}, cr, time.Now())
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: "credentials", Namespace: victimNamespace}, &corev1.Secret{})).To(Succeed())
			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: "cluster-admin-like"}, &rbacv1.ClusterRole{})).To(Succeed())

			err = fakeClient.Get(ctx, types.NamespacedName{Name: "credentials", Namespace: tenantNamespace}, &corev1.Secret{})
			Expect(apierrors.IsNotFound(err)).To(BeTrue(), "own-namespace secret should be cleaned up")
		})
	})
})

var _ = Describe("Timezone validation", func() {
	DescribeTable("validateTimezone",
		func(tz string, valid bool) {
			err := (&CronJobScaleDownReconciler{}).validateTimezone(tz)
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}
		},
		Entry("UTC", "UTC", true),
		Entry("region/city", "America/New_York", true),
		Entry("three-level zone", "America/Argentina/Buenos_Aires", true),
		Entry("Etc offset", "Etc/GMT+5", true),
		Entry("tzdb name with digits", "EST5EDT", true),
		Entry("tzdb name with digits", "PST8PDT", true),
		Entry("path traversal", "../etc/passwd", false),
		Entry("embedded dot segment", "UTC/../x", false),
		Entry("leading slash", "/UTC", false),
		Entry("unknown zone", "Mars/Olympus_Mons", false),
	)
})
