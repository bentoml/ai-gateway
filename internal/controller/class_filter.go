// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// newManagedClasses turns the configured GatewayClass name list into a set for
// cheap membership checks. A nil or empty input returns nil, which the helpers
// below interpret as "unfiltered" (backward compatible cluster-wide mode).
func newManagedClasses(names []string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(names))
	for _, n := range names {
		m[n] = struct{}{}
	}
	return m
}

// gatewayInManagedClass reports whether a Gateway's spec.gatewayClassName is in
// the managed set. An empty set means unfiltered (always true). A nil Gateway
// means unmanaged.
func gatewayInManagedClass(managed map[string]struct{}, gw *gwapiv1.Gateway) bool {
	if len(managed) == 0 {
		return true
	}
	if gw == nil {
		return false
	}
	_, ok := managed[string(gw.Spec.GatewayClassName)]
	return ok
}

// anyParentGatewayManaged reports whether any Gateway referenced by the given
// parentRefs belongs to a managed class. Unresolvable refs (not-found etc.) and
// non-Gateway kinds are skipped. Returns true unconditionally when the managed
// set is empty.
func anyParentGatewayManaged(ctx context.Context, c client.Client, managed map[string]struct{}, defaultNamespace string, refs []gwapiv1.ParentReference) bool {
	if len(managed) == 0 {
		return true
	}
	for _, p := range refs {
		if p.Kind != nil && *p.Kind != "Gateway" {
			continue
		}
		ns := defaultNamespace
		if p.Namespace != nil {
			ns = string(*p.Namespace)
		}
		var gw gwapiv1.Gateway
		if err := c.Get(ctx, client.ObjectKey{Name: string(p.Name), Namespace: ns}, &gw); err != nil {
			continue
		}
		if gatewayInManagedClass(managed, &gw) {
			return true
		}
	}
	return false
}
