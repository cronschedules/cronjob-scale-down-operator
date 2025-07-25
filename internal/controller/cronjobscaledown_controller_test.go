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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cronschedulesv1 "github.com/z4ck404/cronjob-scale-down-operator/api/v1"
)

var _ = Describe("CronJobScaleDown Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		cronjobscaledown := &cronschedulesv1.CronJobScaleDown{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind CronJobScaleDown")
			err := k8sClient.Get(ctx, typeNamespacedName, cronjobscaledown)
			if err != nil && errors.IsNotFound(err) {
				resource := &cronschedulesv1.CronJobScaleDown{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: cronschedulesv1.CronJobScaleDownSpec{
						TargetRef: &cronschedulesv1.TargetRef{
							Name:      "test-deployment",
							Namespace: "default",
							Kind:      "Deployment",
						},
						TimeZone: "Europe/Berlin",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &cronschedulesv1.CronJobScaleDown{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance CronJobScaleDown")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CronJobScaleDownReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})

	Context("When testing cleanup functionality", func() {
		const cleanupResourceName = "test-cleanup-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      cleanupResourceName,
			Namespace: "default",
		}
		cronjobscaledown := &cronschedulesv1.CronJobScaleDown{}

		BeforeEach(func() {
			By("creating the custom resource for cleanup testing")
			err := k8sClient.Get(ctx, typeNamespacedName, cronjobscaledown)
			if err != nil && errors.IsNotFound(err) {
				resource := &cronschedulesv1.CronJobScaleDown{
					ObjectMeta: metav1.ObjectMeta{
						Name:      cleanupResourceName,
						Namespace: "default",
					},
					Spec: cronschedulesv1.CronJobScaleDownSpec{
						CleanupSchedule: "*/30 * * * * *", // Every 30 seconds for testing
						CleanupConfig: &cronschedulesv1.CleanupConfig{
							AnnotationKey: "test.elbazi.co/cleanup-after",
							ResourceTypes: []string{"ConfigMap"},
							DryRun:        true, // Use dry run for testing
						},
						TimeZone: "UTC",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &cronschedulesv1.CronJobScaleDown{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance CronJobScaleDown")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should reconcile without error for cleanup-only configuration", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CronJobScaleDownReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
