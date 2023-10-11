/*
Copyright 2023 The Kubernetes Authors.

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
package workload

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	nodev1 "k8s.io/api/node/v1"

	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
	utiltesting "sigs.k8s.io/kueue/pkg/util/testing"
)

func TestAdjustResources(t *testing.T) {
	cases := map[string]struct {
		runtimeClasses []nodev1.RuntimeClass
		limitranges    []corev1.LimitRange
		wl             *kueue.Workload
		wantWl         *kueue.Workload
	}{
		"Limits applied to requests": {
			wl: utiltesting.MakeWorkload("foo", "").
				PodSets(
					*utiltesting.MakePodSet("a", 1).
						Limit(corev1.ResourceCPU, "1").
						Limit(corev1.ResourceMemory, "1Gi").
						Obj(),
					*utiltesting.MakePodSet("b", 1).
						Request(corev1.ResourceCPU, "2").
						Limit(corev1.ResourceCPU, "3").
						Limit(corev1.ResourceMemory, "1Gi").
						Obj(),
				).
				Obj(),
			wantWl: utiltesting.MakeWorkload("foo", "").
				PodSets(
					*utiltesting.MakePodSet("a", 1).
						Limit(corev1.ResourceCPU, "1").
						Limit(corev1.ResourceMemory, "1Gi").
						Request(corev1.ResourceCPU, "1").
						Request(corev1.ResourceMemory, "1Gi").
						Obj(),
					*utiltesting.MakePodSet("b", 1).
						Limit(corev1.ResourceCPU, "2").
						Limit(corev1.ResourceMemory, "1Gi").
						Request(corev1.ResourceCPU, "3").
						Request(corev1.ResourceMemory, "1Gi").
						Obj(),
				).
				Obj(),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cl := utiltesting.NewClientBuilder().WithLists(
				&nodev1.RuntimeClassList{Items: tc.runtimeClasses},
				&corev1.LimitRangeList{Items: tc.limitranges},
			).Build()
			ctx, _ := utiltesting.ContextWithLog(t)
			AdjustResources(ctx, cl, tc.wl)
			if diff := cmp.Diff(tc.wl, tc.wantWl); diff != "" {
				t.Error("Unexpected resources after adjusting (-want,+got): %s", diff)
			}
		})
	}
}
