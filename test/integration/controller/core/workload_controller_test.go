/*
Copyright 2022 The Kubernetes Authors.

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

package core

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	nodev1 "k8s.io/api/node/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1alpha1"
	"sigs.k8s.io/kueue/pkg/util/pointer"
	"sigs.k8s.io/kueue/pkg/util/testing"
	"sigs.k8s.io/kueue/pkg/workload"
	"sigs.k8s.io/kueue/test/integration/framework"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = ginkgo.Describe("Workload controller", func() {
	var (
		ns                   *corev1.Namespace
		updatedQueueWorkload kueue.Workload
		queue                *kueue.Queue
		wl                   *kueue.Workload
		message              string
		runtimeClass         *nodev1.RuntimeClass
		clusterQueue         *kueue.ClusterQueue
		updatedCQ            kueue.ClusterQueue
		resources            = corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("1"),
		}
	)

	ginkgo.BeforeEach(func() {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "core-workload-",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, ns)).To(gomega.Succeed())
	})

	ginkgo.AfterEach(func() {
		clusterQueue = nil
		queue = nil
		updatedQueueWorkload = kueue.Workload{}
		updatedCQ = kueue.ClusterQueue{}
	})

	ginkgo.When("the queue is not defined in the workload", func() {
		ginkgo.AfterEach(func() {
			gomega.Expect(framework.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
		})
		ginkgo.It("Should update status when workloads are created", func() {
			wl = testing.MakeWorkload("one", ns.Name).Request(corev1.ResourceCPU, "1").Obj()
			message = fmt.Sprintf("Queue %s doesn't exist", "")
			gomega.Expect(k8sClient.Create(ctx, wl)).To(gomega.Succeed())
			gomega.Eventually(func() int {
				gomega.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(wl), &updatedQueueWorkload)).To(gomega.Succeed())
				return len(updatedQueueWorkload.Status.Conditions)
			}, framework.Timeout, framework.Interval).Should(gomega.BeComparableTo(1))
			gomega.Expect(updatedQueueWorkload.Status.Conditions[0].Message).To(gomega.BeComparableTo(message))
		})
	})

	ginkgo.When("the queue doesn't exist", func() {
		ginkgo.AfterEach(func() {
			gomega.Expect(framework.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
		})
		ginkgo.It("Should update status when workloads are created", func() {
			wl = testing.MakeWorkload("two", ns.Name).Queue("nonCreatedQueue").Request(corev1.ResourceCPU, "1").Obj()
			message = fmt.Sprintf("Queue %s doesn't exist", "nonCreatedQueue")
			gomega.Expect(k8sClient.Create(ctx, wl)).To(gomega.Succeed())
			gomega.Eventually(func() int {
				gomega.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(wl), &updatedQueueWorkload)).To(gomega.Succeed())
				return len(updatedQueueWorkload.Status.Conditions)
			}, framework.Timeout, framework.Interval).Should(gomega.BeComparableTo(1))
			gomega.Expect(updatedQueueWorkload.Status.Conditions[0].Message).To(gomega.BeComparableTo(message))
		})
	})

	ginkgo.When("the clusterqueue doesn't exist", func() {
		ginkgo.BeforeEach(func() {
			queue = testing.MakeQueue("queue", ns.Name).ClusterQueue("fooclusterqueue").Obj()
			gomega.Expect(k8sClient.Create(ctx, queue)).To(gomega.Succeed())
		})
		ginkgo.AfterEach(func() {
			gomega.Expect(framework.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
		})
		ginkgo.It("Should update status when workloads are created", func() {
			wl = testing.MakeWorkload("three", ns.Name).Queue(queue.Name).Request(corev1.ResourceCPU, "1").Obj()
			message = fmt.Sprintf("ClusterQueue %s doesn't exist", "fooclusterqueue")
			gomega.Expect(k8sClient.Create(ctx, wl)).To(gomega.Succeed())
			gomega.Eventually(func() []kueue.WorkloadCondition {
				gomega.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(wl), &updatedQueueWorkload)).To(gomega.Succeed())
				return updatedQueueWorkload.Status.Conditions
			}, framework.Timeout, framework.Interval).ShouldNot(gomega.BeNil())
			gomega.Expect(updatedQueueWorkload.Status.Conditions[0].Message).To(gomega.BeComparableTo(message))
		})
	})

	ginkgo.When("the workload is admitted", func() {
		var flavor *kueue.ResourceFlavor

		ginkgo.BeforeEach(func() {
			flavor = testing.MakeResourceFlavor(flavorOnDemand).Obj()
			gomega.Expect(k8sClient.Create(ctx, flavor)).Should(gomega.Succeed())
			clusterQueue = testing.MakeClusterQueue("cluster-queue").
				Resource(testing.MakeResource(resourceGPU).
					Flavor(testing.MakeFlavor(flavorOnDemand, "5").Max("10").Obj()).Obj()).
				Obj()
			gomega.Expect(k8sClient.Create(ctx, clusterQueue)).To(gomega.Succeed())
			queue = testing.MakeQueue("queue", ns.Name).ClusterQueue(clusterQueue.Name).Obj()
			gomega.Expect(k8sClient.Create(ctx, queue)).To(gomega.Succeed())
		})
		ginkgo.AfterEach(func() {
			gomega.Expect(framework.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
			gomega.Expect(framework.DeleteResourceFlavor(ctx, k8sClient, flavor)).To(gomega.Succeed())
			gomega.Expect(framework.DeleteClusterQueue(ctx, k8sClient, clusterQueue)).To(gomega.Succeed())
		})

		ginkgo.It("Should update the workload's condition", func() {
			ginkgo.By("Create workload")
			wl = testing.MakeWorkload("one", ns.Name).Queue(queue.Name).Request(corev1.ResourceCPU, "1").Obj()
			gomega.Expect(k8sClient.Create(ctx, wl)).To(gomega.Succeed())

			ginkgo.By("Admit workload")
			gomega.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(wl), &updatedQueueWorkload)).To(gomega.Succeed())
			updatedQueueWorkload.Spec.Admission = testing.MakeAdmission(clusterQueue.Name).
				Flavor(corev1.ResourceCPU, flavorOnDemand).Obj()
			gomega.Expect(k8sClient.Update(ctx, &updatedQueueWorkload)).To(gomega.Succeed())
			gomega.Eventually(func() bool {
				gomega.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(wl), &updatedQueueWorkload)).To(gomega.Succeed())
				return workload.InCondition(&updatedQueueWorkload, kueue.WorkloadAdmitted)
			}, framework.Timeout, framework.Interval).Should(gomega.BeTrue())
		})
	})

	ginkgo.When("Workload with RuntimeClass defined", func() {
		ginkgo.BeforeEach(func() {
			runtimeClass = testing.MakeRuntimeClass("kata", "bar-handler").PodOverhead(resources).Obj()
			gomega.Expect(k8sClient.Create(ctx, runtimeClass)).To(gomega.Succeed())
			clusterQueue = testing.MakeClusterQueue("clusterqueue").
				Resource(testing.MakeResource(corev1.ResourceCPU).
					Flavor(testing.MakeFlavor(flavorOnDemand, "5").Max("10").Obj()).Obj()).
				Obj()
			gomega.Expect(k8sClient.Create(ctx, clusterQueue)).To(gomega.Succeed())
			queue = testing.MakeQueue("queue", ns.Name).ClusterQueue(clusterQueue.Name).Obj()
			gomega.Expect(k8sClient.Create(ctx, queue)).To(gomega.Succeed())
		})
		ginkgo.AfterEach(func() {
			gomega.Expect(framework.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
			gomega.Expect(framework.DeleteRuntimeClass(ctx, k8sClient, runtimeClass)).To(gomega.Succeed())
			framework.ExpectClusterQueueToBeDeleted(ctx, k8sClient, clusterQueue, true)
		})

		ginkgo.It("Should accumulate RuntimeClass's overhead", func() {
			ginkgo.By("Create workload")
			wl = testing.MakeWorkload("one", ns.Name).
				Queue(queue.Name).
				Request(corev1.ResourceCPU, "1").
				Admit(testing.MakeAdmission(clusterQueue.Name).
					Flavor(corev1.ResourceCPU, flavorOnDemand).Obj()).
				RuntimeClass("kata").
				Obj()
			gomega.Expect(k8sClient.Create(ctx, wl)).To(gomega.Succeed())

			ginkgo.By("Got ClusterQueueStatus")
			gomega.Eventually(func() kueue.ClusterQueueStatus {
				gomega.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterQueue), &updatedCQ)).To(gomega.Succeed())
				return updatedCQ.Status
			}, framework.Timeout, framework.Interval).Should(gomega.BeComparableTo(kueue.ClusterQueueStatus{
				PendingWorkloads:  0,
				AdmittedWorkloads: 1,
				UsedResources: kueue.UsedResources{
					corev1.ResourceCPU: {
						flavorOnDemand: {
							Total:    pointer.Quantity(resource.MustParse("2")),
							Borrowed: nil,
						},
					},
				},
			}))
		})
	})

	ginkgo.When("Workload with non-existent RuntimeClass defined", func() {
		ginkgo.BeforeEach(func() {
			clusterQueue = testing.MakeClusterQueue("clusterqueue").
				Resource(testing.MakeResource(corev1.ResourceCPU).
					Flavor(testing.MakeFlavor(flavorOnDemand, "5").Max("10").Obj()).Obj()).
				Obj()
			gomega.Expect(k8sClient.Create(ctx, clusterQueue)).To(gomega.Succeed())
			queue = testing.MakeQueue("queue", ns.Name).ClusterQueue(clusterQueue.Name).Obj()
			gomega.Expect(k8sClient.Create(ctx, queue)).To(gomega.Succeed())
		})
		ginkgo.AfterEach(func() {
			gomega.Expect(framework.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
			framework.ExpectClusterQueueToBeDeleted(ctx, k8sClient, clusterQueue, true)
		})

		ginkgo.It("Should not accumulate RuntimeClass's overhead", func() {
			ginkgo.By("Create workload")
			wl = testing.MakeWorkload("one", ns.Name).
				Queue(queue.Name).
				Request(corev1.ResourceCPU, "1").
				Admit(testing.MakeAdmission(clusterQueue.Name).
					Flavor(corev1.ResourceCPU, flavorOnDemand).Obj()).
				RuntimeClass("kata").
				Obj()
			gomega.Expect(k8sClient.Create(ctx, wl)).To(gomega.Succeed())

			ginkgo.By("Got ClusterQueueStatus")
			gomega.Eventually(func() kueue.ClusterQueueStatus {
				gomega.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterQueue), &updatedCQ)).To(gomega.Succeed())
				return updatedCQ.Status
			}, framework.Timeout, framework.Interval).Should(gomega.BeComparableTo(kueue.ClusterQueueStatus{
				PendingWorkloads:  0,
				AdmittedWorkloads: 1,
				UsedResources: kueue.UsedResources{
					corev1.ResourceCPU: {
						flavorOnDemand: {
							Total:    pointer.Quantity(resource.MustParse("1")),
							Borrowed: nil,
						},
					},
				},
			}))
		})
	})
})
