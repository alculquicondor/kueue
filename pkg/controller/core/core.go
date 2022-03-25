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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/kueue/pkg/cache"
	"sigs.k8s.io/kueue/pkg/queue"
)

// SetupControllers sets up the core controllers. It returns the name of the
// controller that failed to create and an error, if any.
func SetupControllers(mgr ctrl.Manager, qManager *queue.Manager, cc *cache.Cache) (string, error) {
	if err := NewQueueReconciler(mgr.GetClient(), qManager).SetupWithManager(mgr); err != nil {
		return "Queue", err
	}
	cqRec := NewClusterQueueReconciler(mgr.GetClient(), qManager, cc)
	if err := cqRec.SetupWithManager(mgr); err != nil {
		return "ClusterQueue", err
	}
	if err := NewQueuedWorkloadReconciler(mgr.GetClient(), qManager, cc, cqRec).SetupWithManager(mgr); err != nil {
		return "QueuedWorkload", err
	}
	if err := NewResourceFlavorReconciler(cc).SetupWithManager(mgr); err != nil {
		return "ResourceFlavor", err
	}
	return "", nil
}
