// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fake2 "k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	aigv1a1 "github.com/envoyproxy/ai-gateway/api/v1alpha1"
	internaltesting "github.com/envoyproxy/ai-gateway/internal/testing"
)

// routeKey reduces verbosity in tests below.
func routeKey(ns, name string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}}
}

func TestAIGatewayRouteController_ClassFilter_EarlyReturn(t *testing.T) {
	fakeClient := requireNewFakeClientWithIndexes(t)
	eventCh := internaltesting.NewControllerEventChan[*gwapiv1.Gateway]()
	c := NewAIGatewayRouteController(fakeClient, fake2.NewClientset(), ctrl.Log, eventCh.Ch, "/v1")
	c.managedClasses = newManagedClasses([]string{"eg"})

	// Gateway in an unmanaged class.
	require.NoError(t, fakeClient.Create(t.Context(), &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gw-b", Namespace: "default"},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "eg-b"},
	}))
	require.NoError(t, fakeClient.Create(t.Context(), &aigv1a1.AIGatewayRoute{
		ObjectMeta: metav1.ObjectMeta{Name: "route-b", Namespace: "default"},
		Spec: aigv1a1.AIGatewayRouteSpec{
			ParentRefs: []gwapiv1a2.ParentReference{{Name: "gw-b"}},
		},
	}))

	_, err := c.Reconcile(t.Context(), routeKey("default", "route-b"))
	require.NoError(t, err)

	var got aigv1a1.AIGatewayRoute
	require.NoError(t, fakeClient.Get(t.Context(), types.NamespacedName{Namespace: "default", Name: "route-b"}, &got))
	// syncAIGatewayRoute adds a finalizer; early-return must skip that step.
	require.NotContains(t, got.Finalizers, aiGatewayControllerFinalizer)
	require.Empty(t, got.Status.Conditions)
}

func TestAIGatewayRouteController_ClassFilter_MixedParentsProceed(t *testing.T) {
	fakeClient := requireNewFakeClientWithIndexes(t)
	eventCh := internaltesting.NewControllerEventChan[*gwapiv1.Gateway]()
	c := NewAIGatewayRouteController(fakeClient, fake2.NewClientset(), ctrl.Log, eventCh.Ch, "/v1")
	c.managedClasses = newManagedClasses([]string{"eg"})

	require.NoError(t, fakeClient.Create(t.Context(), &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gw-a", Namespace: "default"},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "eg"},
	}))
	require.NoError(t, fakeClient.Create(t.Context(), &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gw-b", Namespace: "default"},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "eg-b"},
	}))
	require.NoError(t, fakeClient.Create(t.Context(), &aigv1a1.AIGatewayRoute{
		ObjectMeta: metav1.ObjectMeta{Name: "route-ab", Namespace: "default"},
		Spec: aigv1a1.AIGatewayRouteSpec{
			ParentRefs: []gwapiv1a2.ParentReference{{Name: "gw-a"}, {Name: "gw-b"}},
		},
	}))

	_, err := c.Reconcile(t.Context(), routeKey("default", "route-ab"))
	require.NoError(t, err)

	var got aigv1a1.AIGatewayRoute
	require.NoError(t, fakeClient.Get(t.Context(), types.NamespacedName{Namespace: "default", Name: "route-ab"}, &got))
	require.Contains(t, got.Finalizers, aiGatewayControllerFinalizer)
}

func TestMCPRouteController_ClassFilter_EarlyReturn(t *testing.T) {
	fakeClient := requireNewFakeClientWithIndexes(t)
	eventCh := internaltesting.NewControllerEventChan[*gwapiv1.Gateway]()
	c := NewMCPRouteController(fakeClient, fake2.NewClientset(), ctrl.Log, eventCh.Ch)
	c.managedClasses = newManagedClasses([]string{"eg"})

	require.NoError(t, fakeClient.Create(t.Context(), &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gw-b", Namespace: "default"},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "eg-b"},
	}))
	require.NoError(t, fakeClient.Create(t.Context(), &aigv1a1.MCPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: "mcp-b", Namespace: "default"},
		Spec: aigv1a1.MCPRouteSpec{
			ParentRefs: []gwapiv1a2.ParentReference{{Name: "gw-b"}},
		},
	}))

	_, err := c.Reconcile(t.Context(), routeKey("default", "mcp-b"))
	require.NoError(t, err)

	var got aigv1a1.MCPRoute
	require.NoError(t, fakeClient.Get(t.Context(), types.NamespacedName{Namespace: "default", Name: "mcp-b"}, &got))
	require.Empty(t, got.Status.Conditions)
}

func TestGatewayMutator_ClassFilter_SkipsUnmanaged(t *testing.T) {
	fakeClient := requireNewFakeClientWithIndexes(t)
	require.NoError(t, fakeClient.Create(t.Context(), &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gw-b", Namespace: "default"},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "eg-b"},
	}))

	g := newTestGatewayMutator(fakeClient, fake2.NewClientset(), nil, nil, nil, nil, "", "", "", false)
	g.managedClasses = newManagedClasses([]string{"eg"})

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gw-b-pod",
			Namespace: "default",
			Labels: map[string]string{
				egOwningGatewayNameLabel:      "gw-b",
				egOwningGatewayNamespaceLabel: "default",
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "envoy"}}},
	}
	require.NoError(t, g.Default(t.Context(), pod))
	// Pod spec must be untouched: no extproc container injected.
	for _, ctr := range pod.Spec.Containers {
		require.NotEqual(t, extProcContainerName, ctr.Name)
	}
	for _, ctr := range pod.Spec.InitContainers {
		require.NotEqual(t, extProcContainerName, ctr.Name)
	}
}
