// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGatewayInManagedClass(t *testing.T) {
	unfiltered := newManagedClasses(nil)
	require.True(t, gatewayInManagedClass(unfiltered, &gwapiv1.Gateway{Spec: gwapiv1.GatewaySpec{GatewayClassName: "eg"}}))
	require.True(t, gatewayInManagedClass(unfiltered, nil), "empty filter short-circuits to true")

	m := newManagedClasses([]string{"eg", "eg-b"})
	require.True(t, gatewayInManagedClass(m, &gwapiv1.Gateway{Spec: gwapiv1.GatewaySpec{GatewayClassName: "eg"}}))
	require.True(t, gatewayInManagedClass(m, &gwapiv1.Gateway{Spec: gwapiv1.GatewaySpec{GatewayClassName: "eg-b"}}))
	require.False(t, gatewayInManagedClass(m, &gwapiv1.Gateway{Spec: gwapiv1.GatewaySpec{GatewayClassName: "eg-c"}}))
	require.False(t, gatewayInManagedClass(m, nil))
}

func TestAnyParentGatewayManaged(t *testing.T) {
	managedGW := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "ns"},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "eg"},
	}
	unmanagedGW := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "ns"},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "eg-b"},
	}
	c := fake.NewClientBuilder().WithScheme(Scheme).WithObjects(managedGW, unmanagedGW).Build()
	ctx := t.Context()

	m := newManagedClasses([]string{"eg"})
	require.True(t, anyParentGatewayManaged(ctx, c, nil, "ns", []gwapiv1.ParentReference{{Name: "b"}}),
		"empty set must short-circuit to true")
	require.True(t, anyParentGatewayManaged(ctx, c, m, "ns", []gwapiv1.ParentReference{{Name: "b"}, {Name: "a"}}),
		"at least one managed parent returns true")
	require.False(t, anyParentGatewayManaged(ctx, c, m, "ns", []gwapiv1.ParentReference{{Name: "b"}}),
		"all parents unmanaged returns false")
	require.False(t, anyParentGatewayManaged(ctx, c, m, "ns", []gwapiv1.ParentReference{{Name: "missing"}, {Name: "b"}}),
		"unresolvable refs skipped; remaining unmanaged yields false")
}

func TestNewManagedClasses(t *testing.T) {
	require.Nil(t, newManagedClasses(nil))
	require.Nil(t, newManagedClasses([]string{}))
	require.Equal(t, map[string]struct{}{"eg": {}}, newManagedClasses([]string{"eg"}))
	require.Equal(t, map[string]struct{}{"eg": {}, "eg-b": {}}, newManagedClasses([]string{"eg", "eg-b"}))
}
