package gwapi

import (
	"context"
	"fmt"

	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Shared Gateway API routing resources (ListenerSet, redirect HTTPRoute,
// Certificate) live in istio-gateways and are contributed to by Routing objects
// in many namespaces. Kubernetes ownerReferences cannot cross namespaces, so the
// shared resources cannot be garbage-collected by owner refs. Instead a single
// membership ConfigMap per hostname tracks the live contributors: each contributor
// registers itself on reconcile and deregisters via its finalizer, and the shared
// resources are deleted once the membership is empty. This is a precise ref-count
// — no cluster-wide Routing scan and no peer-spec parsing.

// SharedMembershipName returns the name of the membership ConfigMap for hostname.
func SharedMembershipName(hostname string) string {
	return fmt.Sprintf("shared-routing-members-%x", util.GenerateHashFromName(hostname))
}

// contributorKey is the ConfigMap data key for one contributing Routing.
// Namespaces are DNS labels (no dots), so "<namespace>.<name>" is unambiguous and
// is a valid ConfigMap data key.
func contributorKey(contributor types.NamespacedName) string {
	return contributor.Namespace + "." + contributor.Name
}

// RegisterSharedContributor records that contributor uses the shared resources
// for hostname, creating the membership ConfigMap if it does not exist yet.
// Idempotent.
func RegisterSharedContributor(ctx context.Context, c client.Client, hostname string, contributor types.NamespacedName) error {
	key := types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cm := &corev1.ConfigMap{}
		err := c.Get(ctx, key, cm)
		if apierrors.IsNotFound(err) {
			created := newMembershipConfigMap(hostname)
			created.Data = map[string]string{contributorKey(contributor): ""}
			if createErr := c.Create(ctx, created); createErr == nil {
				return nil
			} else if !apierrors.IsAlreadyExists(createErr) {
				return createErr
			}
			// Lost the create race — re-read and fall through to update.
			if err = c.Get(ctx, key, cm); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if _, ok := cm.Data[contributorKey(contributor)]; ok {
			return nil
		}
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[contributorKey(contributor)] = ""
		return c.Update(ctx, cm)
	})
}

// DeregisterSharedContributor removes contributor from hostname's membership.
// Returns empty=true and deletes the membership ConfigMap when no contributors
// remain, closing the TOCTOU window between checking empty and deletion. The
// deletion is preconditioned on the ResourceVersion from the read; a concurrent
// RegisterSharedContributor that inserts between the read and delete will cause
// a Conflict that RetryOnConflict resolves by re-reading the updated membership.
// A missing ConfigMap is treated as "no contributors remain".
func DeregisterSharedContributor(ctx context.Context, c client.Client, hostname string, contributor types.NamespacedName) (empty bool, err error) {
	key := types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cm := &corev1.ConfigMap{}
		getErr := c.Get(ctx, key, cm)
		if apierrors.IsNotFound(getErr) {
			empty = true
			return nil
		}
		if getErr != nil {
			return getErr
		}
		delete(cm.Data, contributorKey(contributor))
		if len(cm.Data) == 0 {
			// Delete the ConfigMap here, using the ResourceVersion from the Get
			// as a precondition. If a concurrent Register slips in between, the
			// delete will return Conflict; RetryOnConflict re-reads and finds the
			// new contributor, so empty stays false and the CM is not deleted.
			if deleteErr := c.Delete(ctx, cm); deleteErr != nil && !apierrors.IsNotFound(deleteErr) {
				return deleteErr
			}
			empty = true
			return nil
		}
		empty = false
		return c.Update(ctx, cm)
	})
	return empty, err
}

// DeleteSharedMembership removes the membership ConfigMap for hostname. Idempotent.
func DeleteSharedMembership(ctx context.Context, c client.Client, hostname string) error {
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}}
	if err := c.Delete(ctx, cm); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete shared routing membership %s/%s: %w", cm.Namespace, cm.Name, err)
	}
	return nil
}

func newMembershipConfigMap(hostname string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: IstioGatewayNamespace,
			Name:      SharedMembershipName(hostname),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":        "skiperator",
				"skiperator.kartverket.no/controller": "routing-shared",
				"skiperator.kartverket.no/hostname":   fmt.Sprintf("%x", util.GenerateHashFromName(hostname)),
			},
		},
	}
}
