/*
Copyright 2017 The Kubernetes Authors.

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

package daemon

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

func TestDaemonSetUpdatesPods(t *testing.T) {
	ds := newDaemonSet("foo")
	manager, podControl, _, err := newTestController(ds)
	if err != nil {
		t.Fatalf("error creating DaemonSets controller: %v", err)
	}
	maxUnavailable := 2
	addNodes(manager.nodeStore, 0, 5, nil)
	err = manager.dsStore.Add(ds)
	if err != nil {
		t.Fatal(err)
	}
	err = syncAndValidateDaemonSets(manager, ds, podControl, 5, 0, 0)
	if err != nil {
		t.Error(err)
	}
	markPodsReady(podControl.podStore)

	ds.Spec.Template.Spec.Containers[0].Image = "foo2/bar2"
	ds.Spec.UpdateStrategy.Type = apps.RollingUpdateDaemonSetStrategyType
	intStr := intstr.FromInt(maxUnavailable)
	ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
	err = manager.dsStore.Update(ds)
	if err != nil {
		t.Fatal(err)
	}

	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, maxUnavailable, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, maxUnavailable, 0, 0)
	if err != nil {
		t.Error(err)
	}
	markPodsReady(podControl.podStore)

	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, maxUnavailable, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, maxUnavailable, 0, 0)
	if err != nil {
		t.Error(err)
	}
	markPodsReady(podControl.podStore)

	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 1, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 1, 0, 0)
	if err != nil {
		t.Error(err)
	}
	markPodsReady(podControl.podStore)

	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
}
func TestDaemonSetUpdatesSaveOldHealthyPods(t *testing.T) {
	ds := newDaemonSet("foo")
	manager, podControl, _, err := newTestController(ds)
	if err != nil {
		t.Fatalf("error creating DaemonSets controller: %v", err)
	}
	addNodes(manager.nodeStore, 0, 20, nil)
	err = manager.dsStore.Add(ds)
	if err != nil {
		t.Fatal(err)
	}
	// expectSyncDaemonSets(t, manager, ds, podControl, 20, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 20, 0, 0)
	if err != nil {
		t.Error(err)
	}
	markPodsReady(podControl.podStore)

	t.Logf("first update to get 10 old pods which should never be touched")
	ds.Spec.Template.Spec.Containers[0].Image = "foo2/bar2"
	ds.Spec.UpdateStrategy.Type = apps.RollingUpdateDaemonSetStrategyType
	maxUnavailable := 10
	intStr := intstr.FromInt(maxUnavailable)
	ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
	err = manager.dsStore.Update(ds)
	if err != nil {
		t.Fatal(err)
	}

	clearExpectations(t, manager, ds, podControl)
	// expectSyncDaemonSets(t, manager, ds, podControl, 0, maxUnavailable, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, maxUnavailable, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	// expectSyncDaemonSets(t, manager, ds, podControl, maxUnavailable, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, maxUnavailable, 0, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	// expectSyncDaemonSets(t, manager, ds, podControl, 0, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)

	// save the pods we want to maintain running
	oldReadyPods := []string{}
	for _, obj := range podControl.podStore.List() {
		pod := obj.(*v1.Pod)
		if podutil.IsPodReady(pod) {
			oldReadyPods = append(oldReadyPods, pod.Name)
		}
	}

	for i := 0; i < 10; i++ {
		maxUnavailable := rand.Intn(10)
		t.Logf("%d iteration, maxUnavailable=%d", i+1, maxUnavailable)
		intStr = intstr.FromInt(maxUnavailable)
		ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
		ds.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("foo2/bar3-%d", i)
		err = manager.dsStore.Update(ds)
		if err != nil {
			t.Fatal(err)
		}

		// only the 10 unavailable pods will be allowed to be updated
		clearExpectations(t, manager, ds, podControl)
		// expectSyncDaemonSets(t, manager, ds, podControl, 0, 10, 0)
		err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 10, 0)
		if err != nil {
			t.Error(err)
		}

		clearExpectations(t, manager, ds, podControl)
		// expectSyncDaemonSets(t, manager, ds, podControl, 10, 0, 0)
		err = syncAndValidateDaemonSets(manager, ds, podControl, 10, 0, 0)
		if err != nil {
			t.Error(err)
		}

		clearExpectations(t, manager, ds, podControl)
		// expectSyncDaemonSets(t, manager, ds, podControl, 0, 0, 0)
		err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
		if err != nil {
			t.Error(err)
		}
		clearExpectations(t, manager, ds, podControl)

		// verify that the ready pods are never touched
		readyPods := []string{}
		t.Logf("looking for %s", strings.Join(oldReadyPods, ", "))
		for _, obj := range podControl.podStore.List() {
			pod := obj.(*v1.Pod)
			if podutil.IsPodReady(pod) {
				readyPods = append(readyPods, pod.Name)
			}
		}
		for _, oldPod := range oldReadyPods {
			if !slicesContains(readyPods, oldPod) {
				t.Errorf("%s has changed in %d-th iteration", oldPod, i)
			}
		}
	}

	maxUnavailable = 11
	intStr = intstr.FromInt(maxUnavailable)
	ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
	ds.Spec.Template.Spec.Containers[0].Image = "foo2/bar4"
	err = manager.dsStore.Update(ds)
	if err != nil {
		t.Fatal(err)
	}

	clearExpectations(t, manager, ds, podControl)
	// expectSyncDaemonSets(t, manager, ds, podControl, 0, maxUnavailable, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, maxUnavailable, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	// expectSyncDaemonSets(t, manager, ds, podControl, maxUnavailable, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, maxUnavailable, 0, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	// expectSyncDaemonSets(t, manager, ds, podControl, 0, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)

	// verify that the ready pods are never touched
	readyPods := []string{}
	for _, obj := range podControl.podStore.List() {
		pod := obj.(*v1.Pod)
		if podutil.IsPodReady(pod) {
			readyPods = append(readyPods, pod.Name)
		}
	}
	if len(readyPods) != 9 {
		t.Errorf("readyPods are different than expected, should be 9 but is %s", strings.Join(readyPods, ", "))
	}
}

func slicesContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestDaemonSetUpdatesAllOldNotReadyPodsAndNewNotReadyPods(t *testing.T) {
	ds := newDaemonSet("foo")
	manager, podControl, _, err := newTestController(ds)
	if err != nil {
		t.Fatalf("error creating DaemonSets controller: %v", err)
	}
	addNodes(manager.nodeStore, 0, 100, nil)
	err = manager.dsStore.Add(ds)
	if err != nil {
		t.Fatal(err)
	}
	// expectSyncDaemonSets(t, manager, ds, podControl, 100, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 100, 0, 0)
	if err != nil {
		t.Error(err)
	}
	markPodsReady(podControl.podStore)
	var hash1 string
	// at this point we have 100 pods runing from daemonset, and we mark
	// the controller has which will be used later on to fake old pods
	for _, obj := range podControl.podStore.List() {
		pod := obj.(*v1.Pod)
		hash1 = pod.Labels[apps.ControllerRevisionHashLabelKey]
		break
	}

	ds.Spec.Template.Spec.Containers[0].Image = "foo2/bar2"
	ds.Spec.UpdateStrategy.Type = apps.RollingUpdateDaemonSetStrategyType
	maxUnavailable := 10
	intStr := intstr.FromInt(maxUnavailable)
	ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
	err = manager.dsStore.Update(ds)
	if err != nil {
		t.Fatal(err)
	}
	// we need to iterate 10 times, since we allow 10 max unavailable, to reach 100 nodes rollout
	for i := 0; i < 10; i++ {
		clearExpectations(t, manager, ds, podControl)
		// expectSyncDaemonSets(t, manager, ds, podControl, 0, maxUnavailable, 0)
		err = syncAndValidateDaemonSets(manager, ds, podControl, 0, maxUnavailable, 0)
		if err != nil {
			t.Error(err)
		}

		clearExpectations(t, manager, ds, podControl)
		// expectSyncDaemonSets(t, manager, ds, podControl, maxUnavailable, 0, 0)
		err = syncAndValidateDaemonSets(manager, ds, podControl, maxUnavailable, 0, 0)
		if err != nil {
			t.Error(err)
		}
		// make sure to mark the pods ready, otherwise the followup rollouts will fail
		markPodsReady(podControl.podStore)
	}

	clearExpectations(t, manager, ds, podControl)
	// expectSyncDaemonSets(t, manager, ds, podControl, 0, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)

	// to reach the following situation
	// - maxUnavailable 10
	// - 88 unavailable new pods
	// - 2 unavailable old pods
	// - 10 available old pods
	oldUnavailablePods := []string{}
	for i, obj := range podControl.podStore.List() {
		pod := obj.(*v1.Pod)
		// mark the latter 90 pods not ready
		if i >= 10 {
			condition := v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionFalse}
			podutil.UpdatePodCondition(&pod.Status, &condition)
		}
		// mark the first 12 pods with older hash
		if i < 12 {
			pod.Labels[apps.ControllerRevisionHashLabelKey] = hash1
			// note down 2 not available old pods
			if i >= 10 {
				oldUnavailablePods = append(oldUnavailablePods, pod.Name)
			}
		}
	}

	clearExpectations(t, manager, ds, podControl)
	t.Logf("expect 10 old pods deletion in 1st iteration")
	// expectSyncDaemonSets(t, manager, ds, podControl, 0, 2, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 2, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	t.Logf("expect 10 new pods creation in 2nd iteration")
	// expectSyncDaemonSets(t, manager, ds, podControl, 2, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 2, 0, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	t.Logf("expect no modifications in 3rd iteration")
	// expectSyncDaemonSets(t, manager, ds, podControl, 0, 0, 0)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)

	// check if oldUnavailablePods were replaced
	t.Logf("Looking for old pods %s", strings.Join(oldUnavailablePods, ", "))
	notUpdatedOldPods := []string{}
	for _, obj := range podControl.podStore.List() {
		pod := obj.(*v1.Pod)
		for _, oldPod := range oldUnavailablePods {
			if pod.Name == oldPod {
				notUpdatedOldPods = append(notUpdatedOldPods, pod.Name)
			}
		}
	}
	if len(notUpdatedOldPods) > 0 {
		t.Fatalf("found not updated old pods: %s", strings.Join(notUpdatedOldPods, ", "))
	}
}

func TestDaemonSetUpdatesWhenNewPosIsNotReady(t *testing.T) {
	ds := newDaemonSet("foo")
	manager, podControl, _, err := newTestController(ds)
	if err != nil {
		t.Fatalf("error creating DaemonSets controller: %v", err)
	}
	maxUnavailable := 3
	addNodes(manager.nodeStore, 0, 5, nil)
	err = manager.dsStore.Add(ds)
	if err != nil {
		t.Fatal(err)
	}
	err = syncAndValidateDaemonSets(manager, ds, podControl, 5, 0, 0)
	if err != nil {
		t.Error(err)
	}
	markPodsReady(podControl.podStore)

	ds.Spec.Template.Spec.Containers[0].Image = "foo2/bar2"
	ds.Spec.UpdateStrategy.Type = apps.RollingUpdateDaemonSetStrategyType
	intStr := intstr.FromInt(maxUnavailable)
	ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
	err = manager.dsStore.Update(ds)
	if err != nil {
		t.Fatal(err)
	}

	// new pods are not ready numUnavailable == maxUnavailable
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, maxUnavailable, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, maxUnavailable, 0, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
}

func TestDaemonSetUpdatesAllOldPodsNotReady(t *testing.T) {
	ds := newDaemonSet("foo")
	manager, podControl, _, err := newTestController(ds)
	if err != nil {
		t.Fatalf("error creating DaemonSets controller: %v", err)
	}
	maxUnavailable := 3
	addNodes(manager.nodeStore, 0, 5, nil)
	err = manager.dsStore.Add(ds)
	if err != nil {
		t.Fatal(err)
	}
	err = syncAndValidateDaemonSets(manager, ds, podControl, 5, 0, 0)
	if err != nil {
		t.Error(err)
	}

	ds.Spec.Template.Spec.Containers[0].Image = "foo2/bar2"
	ds.Spec.UpdateStrategy.Type = apps.RollingUpdateDaemonSetStrategyType
	intStr := intstr.FromInt(maxUnavailable)
	ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
	err = manager.dsStore.Update(ds)
	if err != nil {
		t.Fatal(err)
	}

	// all old pods are unavailable so should be removed
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 5, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 5, 0, 0)
	if err != nil {
		t.Error(err)
	}

	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
}

func TestDaemonSetUpdatesNoTemplateChanged(t *testing.T) {
	ds := newDaemonSet("foo")
	manager, podControl, _, err := newTestController(ds)
	if err != nil {
		t.Fatalf("error creating DaemonSets controller: %v", err)
	}
	maxUnavailable := 3
	addNodes(manager.nodeStore, 0, 5, nil)
	err = manager.dsStore.Add(ds)
	if err != nil {
		t.Fatal(err)
	}
	err = syncAndValidateDaemonSets(manager, ds, podControl, 5, 0, 0)
	if err != nil {
		t.Error(err)
	}

	ds.Spec.UpdateStrategy.Type = apps.RollingUpdateDaemonSetStrategyType
	intStr := intstr.FromInt(maxUnavailable)
	ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
	err = manager.dsStore.Update(ds)
	if err != nil {
		t.Fatal(err)
	}

	// template is not changed no pod should be removed
	clearExpectations(t, manager, ds, podControl)
	err = syncAndValidateDaemonSets(manager, ds, podControl, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	clearExpectations(t, manager, ds, podControl)
}

func TestGetUnavailableNumbers(t *testing.T) {
	cases := []struct {
		name           string
		Manager        *daemonSetsController
		ds             *apps.DaemonSet
		nodeToPods     map[string][]*v1.Pod
		maxUnavailable int
		numUnavailable int
		Err            error
	}{
		{
			name: "No nodes",
			Manager: func() *daemonSetsController {
				manager, _, _, err := newTestController()
				if err != nil {
					t.Fatalf("error creating DaemonSets controller: %v", err)
				}
				return manager
			}(),
			ds: func() *apps.DaemonSet {
				ds := newDaemonSet("x")
				intStr := intstr.FromInt(0)
				ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
				return ds
			}(),
			nodeToPods:     make(map[string][]*v1.Pod),
			maxUnavailable: 0,
			numUnavailable: 0,
		},
		{
			name: "Two nodes with ready pods",
			Manager: func() *daemonSetsController {
				manager, _, _, err := newTestController()
				if err != nil {
					t.Fatalf("error creating DaemonSets controller: %v", err)
				}
				addNodes(manager.nodeStore, 0, 2, nil)
				return manager
			}(),
			ds: func() *apps.DaemonSet {
				ds := newDaemonSet("x")
				intStr := intstr.FromInt(1)
				ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
				return ds
			}(),
			nodeToPods: func() map[string][]*v1.Pod {
				mapping := make(map[string][]*v1.Pod)
				pod0 := newPod("pod-0", "node-0", simpleDaemonSetLabel, nil)
				pod1 := newPod("pod-1", "node-1", simpleDaemonSetLabel, nil)
				markPodReady(pod0)
				markPodReady(pod1)
				mapping["node-0"] = []*v1.Pod{pod0}
				mapping["node-1"] = []*v1.Pod{pod1}
				return mapping
			}(),
			maxUnavailable: 1,
			numUnavailable: 0,
		},
		{
			name: "Two nodes, one node without pods",
			Manager: func() *daemonSetsController {
				manager, _, _, err := newTestController()
				if err != nil {
					t.Fatalf("error creating DaemonSets controller: %v", err)
				}
				addNodes(manager.nodeStore, 0, 2, nil)
				return manager
			}(),
			ds: func() *apps.DaemonSet {
				ds := newDaemonSet("x")
				intStr := intstr.FromInt(0)
				ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
				return ds
			}(),
			nodeToPods: func() map[string][]*v1.Pod {
				mapping := make(map[string][]*v1.Pod)
				pod0 := newPod("pod-0", "node-0", simpleDaemonSetLabel, nil)
				markPodReady(pod0)
				mapping["node-0"] = []*v1.Pod{pod0}
				return mapping
			}(),
			maxUnavailable: 0,
			numUnavailable: 1,
		},
		{
			name: "Two nodes with pods, MaxUnavailable in percents",
			Manager: func() *daemonSetsController {
				manager, _, _, err := newTestController()
				if err != nil {
					t.Fatalf("error creating DaemonSets controller: %v", err)
				}
				addNodes(manager.nodeStore, 0, 2, nil)
				return manager
			}(),
			ds: func() *apps.DaemonSet {
				ds := newDaemonSet("x")
				intStr := intstr.FromString("50%")
				ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
				return ds
			}(),
			nodeToPods: func() map[string][]*v1.Pod {
				mapping := make(map[string][]*v1.Pod)
				pod0 := newPod("pod-0", "node-0", simpleDaemonSetLabel, nil)
				pod1 := newPod("pod-1", "node-1", simpleDaemonSetLabel, nil)
				markPodReady(pod0)
				markPodReady(pod1)
				mapping["node-0"] = []*v1.Pod{pod0}
				mapping["node-1"] = []*v1.Pod{pod1}
				return mapping
			}(),
			maxUnavailable: 1,
			numUnavailable: 0,
		},
		{
			name: "Two nodes with pods, MaxUnavailable in percents, pod terminating",
			Manager: func() *daemonSetsController {
				manager, _, _, err := newTestController()
				if err != nil {
					t.Fatalf("error creating DaemonSets controller: %v", err)
				}
				addNodes(manager.nodeStore, 0, 2, nil)
				return manager
			}(),
			ds: func() *apps.DaemonSet {
				ds := newDaemonSet("x")
				intStr := intstr.FromString("50%")
				ds.Spec.UpdateStrategy.RollingUpdate = &apps.RollingUpdateDaemonSet{MaxUnavailable: &intStr}
				return ds
			}(),
			nodeToPods: func() map[string][]*v1.Pod {
				mapping := make(map[string][]*v1.Pod)
				pod0 := newPod("pod-0", "node-0", simpleDaemonSetLabel, nil)
				pod1 := newPod("pod-1", "node-1", simpleDaemonSetLabel, nil)
				now := metav1.Now()
				markPodReady(pod0)
				markPodReady(pod1)
				pod1.DeletionTimestamp = &now
				mapping["node-0"] = []*v1.Pod{pod0}
				mapping["node-1"] = []*v1.Pod{pod1}
				return mapping
			}(),
			maxUnavailable: 1,
			numUnavailable: 1,
		},
	}

	for _, c := range cases {
		c.Manager.dsStore.Add(c.ds)
		nodeList, err := c.Manager.nodeLister.List(labels.Everything())
		if err != nil {
			t.Fatalf("error listing nodes: %v", err)
		}
		maxUnavailable, numUnavailable, err := c.Manager.getUnavailableNumbers(c.ds, nodeList, c.nodeToPods)
		if err != nil && c.Err != nil {
			if c.Err != err {
				t.Errorf("Test case: %s. Expected error: %v but got: %v", c.name, c.Err, err)
			}
		} else if err != nil {
			t.Errorf("Test case: %s. Unexpected error: %v", c.name, err)
		} else if maxUnavailable != c.maxUnavailable || numUnavailable != c.numUnavailable {
			t.Errorf("Test case: %s. Wrong values. maxUnavailable: %d, expected: %d, numUnavailable: %d. expected: %d", c.name, maxUnavailable, c.maxUnavailable, numUnavailable, c.numUnavailable)
		}
	}
}
