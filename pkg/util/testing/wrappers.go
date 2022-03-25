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

package testing

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kueue "sigs.k8s.io/kueue/api/v1alpha1"
	"sigs.k8s.io/kueue/pkg/constants"
	"sigs.k8s.io/kueue/pkg/util/pointer"
)

// JobWrapper wraps a Job.
type JobWrapper struct{ batchv1.Job }

// MakeJob creates a wrapper for a suspended job with a single container and parallelism=1.
func MakeJob(name, ns string) *JobWrapper {
	return &JobWrapper{batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Annotations: make(map[string]string, 1),
		},
		Spec: batchv1.JobSpec{
			Parallelism: pointer.Int32(1),
			Suspend:     pointer.Bool(true),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:      "c",
							Image:     "pause",
							Command:   []string{},
							Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{}},
						},
					},
					NodeSelector: map[string]string{},
				},
			},
		},
	}}
}

// Obj returns the inner Job.
func (j *JobWrapper) Obj() *batchv1.Job {
	return &j.Job
}

// Suspend updates the suspend status of the job
func (j *JobWrapper) Suspend(s bool) *JobWrapper {
	j.Spec.Suspend = pointer.Bool(s)
	return j
}

// Parallelism updates job parallelism.
func (j *JobWrapper) Parallelism(p int32) *JobWrapper {
	j.Spec.Parallelism = pointer.Int32(p)
	return j
}

// PriorityClass updates job priorityclass.
func (j *JobWrapper) PriorityClass(pc string) *JobWrapper {
	j.Spec.Template.Spec.PriorityClassName = pc
	return j
}

// Queue updates the queue name of the job
func (j *JobWrapper) Queue(queue string) *JobWrapper {
	j.Annotations[constants.QueueAnnotation] = queue
	return j
}

// Toleration adds a toleration to the job.
func (j *JobWrapper) Toleration(t corev1.Toleration) *JobWrapper {
	j.Spec.Template.Spec.Tolerations = append(j.Spec.Template.Spec.Tolerations, t)
	return j
}

// NodeSelector adds a node selector to the job.
func (j *JobWrapper) NodeSelector(k, v string) *JobWrapper {
	j.Spec.Template.Spec.NodeSelector[k] = v
	return j
}

// Request adds a resource request to the default container.
func (j *JobWrapper) Request(r corev1.ResourceName, v string) *JobWrapper {
	j.Spec.Template.Spec.Containers[0].Resources.Requests[r] = resource.MustParse(v)
	return j
}

// PriorityClassWrapper wraps a PriorityClass.
type PriorityClassWrapper struct {
	schedulingv1.PriorityClass
}

// MakePriorityClass creates a wrapper for a PriorityClass.
func MakePriorityClass(name string) *PriorityClassWrapper {
	return &PriorityClassWrapper{schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		}},
	}
}

// PriorityValue update value of PriorityClass。
func (p *PriorityClassWrapper) PriorityValue(v int32) *PriorityClassWrapper {
	p.Value = v
	return p
}

// Obj returns the inner PriorityClass.
func (p *PriorityClassWrapper) Obj() *schedulingv1.PriorityClass {
	return &p.PriorityClass
}

type QueuedWorkloadWrapper struct{ kueue.QueuedWorkload }

// MakeQueuedWorkload creates a wrapper for a QueuedWorkload with a single
// pod with a single container.
func MakeQueuedWorkload(name, ns string) *QueuedWorkloadWrapper {
	return &QueuedWorkloadWrapper{kueue.QueuedWorkload{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: kueue.QueuedWorkloadSpec{
			PodSets: []kueue.PodSet{
				{
					Name:  "main",
					Count: 1,
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "c",
								Resources: corev1.ResourceRequirements{
									Requests: make(corev1.ResourceList),
								},
							},
						},
					},
				},
			},
		},
	}}
}

func (w *QueuedWorkloadWrapper) Obj() *kueue.QueuedWorkload {
	return &w.QueuedWorkload
}

func (w *QueuedWorkloadWrapper) Request(r corev1.ResourceName, q string) *QueuedWorkloadWrapper {
	w.Spec.PodSets[0].Spec.Containers[0].Resources.Requests[r] = resource.MustParse(q)
	return w
}

func (w *QueuedWorkloadWrapper) Queue(q string) *QueuedWorkloadWrapper {
	w.Spec.QueueName = q
	return w
}

func (w *QueuedWorkloadWrapper) Admit(a *kueue.Admission) *QueuedWorkloadWrapper {
	w.Spec.Admission = a
	return w
}

func (w *QueuedWorkloadWrapper) Creation(t time.Time) *QueuedWorkloadWrapper {
	w.CreationTimestamp = metav1.NewTime(t)
	return w
}

func (w *QueuedWorkloadWrapper) PriorityClass(priorityClassName string) *QueuedWorkloadWrapper {
	w.Spec.PriorityClassName = priorityClassName
	return w
}

// AdmissionWrapper wraps an Admission
type AdmissionWrapper struct{ kueue.Admission }

func MakeAdmission(cq string) *AdmissionWrapper {
	return &AdmissionWrapper{kueue.Admission{
		ClusterQueue: kueue.ClusterQueueReference(cq),
		PodSetFlavors: []kueue.PodSetFlavors{
			{
				Name:    "main",
				Flavors: make(map[corev1.ResourceName]string),
			},
		},
	}}
}

func (w *AdmissionWrapper) Obj() *kueue.Admission {
	return &w.Admission
}

func (w *AdmissionWrapper) Flavor(r corev1.ResourceName, f string) *AdmissionWrapper {
	w.PodSetFlavors[0].Flavors[r] = f
	return w
}

// QueueWrapper wraps a Queue.
type QueueWrapper struct{ kueue.Queue }

// MakeQueue creates a wrapper for a Queue.
func MakeQueue(name, ns string) *QueueWrapper {
	return &QueueWrapper{kueue.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}}
}

// Obj returns the inner Queue.
func (q *QueueWrapper) Obj() *kueue.Queue {
	return &q.Queue
}

// ClusterQueue updates the clusterQueue the queue points to.
func (q *QueueWrapper) ClusterQueue(c string) *QueueWrapper {
	q.Spec.ClusterQueue = kueue.ClusterQueueReference(c)
	return q
}

// ClusterQueueWrapper wraps a ClusterQueue.
type ClusterQueueWrapper struct{ kueue.ClusterQueue }

// MakeClusterQueue creates a wrapper for a ClusterQueue with a
// select-all NamespaceSelector.
func MakeClusterQueue(name string) *ClusterQueueWrapper {
	return &ClusterQueueWrapper{kueue.ClusterQueue{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: kueue.ClusterQueueSpec{
			NamespaceSelector: &metav1.LabelSelector{},
			QueueingStrategy:  kueue.StrictFIFO,
		},
	}}
}

// Obj returns the inner ClusterQueue.
func (c *ClusterQueueWrapper) Obj() *kueue.ClusterQueue {
	return &c.ClusterQueue
}

// Cohort sets the borrowing cohort.
func (c *ClusterQueueWrapper) Cohort(cohort string) *ClusterQueueWrapper {
	c.Spec.Cohort = cohort
	return c
}

// Resource adds a resource with flavors.
func (c *ClusterQueueWrapper) Resource(r *kueue.Resource) *ClusterQueueWrapper {
	c.Spec.RequestableResources = append(c.Spec.RequestableResources, *r)
	return c
}

// QueueingStrategy sets the queueing strategy in this ClusterQueue.
func (c *ClusterQueueWrapper) QueueingStrategy(strategy kueue.QueueingStrategy) *ClusterQueueWrapper {
	c.Spec.QueueingStrategy = strategy
	return c
}

// NamespaceSelector sets the namespace selector.
func (c *ClusterQueueWrapper) NamespaceSelector(s *metav1.LabelSelector) *ClusterQueueWrapper {
	c.Spec.NamespaceSelector = s
	return c
}

// ResourceWrapper wraps a requestable resource.
type ResourceWrapper struct{ kueue.Resource }

// MakeResource creates a wrapper for a requestable resource.
func MakeResource(name corev1.ResourceName) *ResourceWrapper {
	return &ResourceWrapper{kueue.Resource{
		Name: name,
	}}
}

// Obj returns the inner resource.
func (r *ResourceWrapper) Obj() *kueue.Resource {
	return &r.Resource
}

// Flavor appends a flavor.
func (r *ResourceWrapper) Flavor(f *kueue.Flavor) *ResourceWrapper {
	r.Flavors = append(r.Flavors, *f)
	return r
}

// FlavorWrapper wraps a resource flavor.
type FlavorWrapper struct{ kueue.Flavor }

// MakeFlavor creates a wrapper for a resource flavor.
func MakeFlavor(rf, guaranteed string) *FlavorWrapper {
	return &FlavorWrapper{kueue.Flavor{
		ResourceFlavor: kueue.ResourceFlavorReference(rf),
		Quota: kueue.Quota{
			Guaranteed: resource.MustParse(guaranteed),
		},
	}}
}

// Obj returns the inner flavor.
func (f *FlavorWrapper) Obj() *kueue.Flavor {
	return &f.Flavor
}

// Ceiling updates the flavor ceiling.
func (f *FlavorWrapper) Ceiling(c string) *FlavorWrapper {
	f.Quota.Ceiling = pointer.Quantity(resource.MustParse(c))
	return f
}

// ResourceFlavorWrapper wraps a ResourceFlavor.
type ResourceFlavorWrapper struct{ kueue.ResourceFlavor }

// MakeResourceFlavor creates a wrapper for a ResourceFlavor.
func MakeResourceFlavor(name string) *ResourceFlavorWrapper {
	return &ResourceFlavorWrapper{kueue.ResourceFlavor{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Labels: map[string]string{},
	}}
}

// Obj returns the inner ResourceFlavor.
func (rf *ResourceFlavorWrapper) Obj() *kueue.ResourceFlavor {
	return &rf.ResourceFlavor
}

// Label adds a label to the ResourceFlavor.
func (rf *ResourceFlavorWrapper) Label(k, v string) *ResourceFlavorWrapper {
	rf.Labels[k] = v
	return rf
}

// Taint adds a taint to the ResourceFlavor.
func (rf *ResourceFlavorWrapper) Taint(t corev1.Taint) *ResourceFlavorWrapper {
	rf.Taints = append(rf.Taints, t)
	return rf
}
