package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func newControllerManagedBy[T client.Object](mgr manager.Manager) wrappedBuilder[T] {
	typ := reflect.ValueOf(*new(T)).Type().Elem()
	obj := reflect.New(typ).Interface().(T)

	return wrappedBuilder[T]{mgr, builder.ControllerManagedBy(mgr).For(obj)}
}

type wrappedBuilder[T client.Object] struct {
	mgr  manager.Manager
	next *builder.Builder
}

func (b wrappedBuilder[T]) Owns(obj client.Object, opts ...builder.OwnsOption) wrappedBuilder[T] {
	return wrappedBuilder[T]{b.mgr, b.next.Owns(obj, opts...)}
}

func (b wrappedBuilder[T]) Watches(src source.Source, handler handler.EventHandler, opts ...builder.WatchesOption) wrappedBuilder[T] {
	return wrappedBuilder[T]{b.mgr, b.next.Watches(src, handler, opts...)}
}

func (b wrappedBuilder[T]) Complete(r objectReconciler[T]) error {
	return b.next.Complete(wrappedReconciler[T]{
		client:   b.mgr.GetClient(),
		recorder: b.mgr.GetEventRecorderFor("skiperator"),
		next:     r,
	})
}

type wrappedReconciler[T client.Object] struct {
	client   client.Client
	recorder record.EventRecorder
	next     objectReconciler[T]
}

type objectReconciler[T client.Object] interface {
	Reconcile(ctx context.Context, obj T) (reconcile.Result, error)
}

func (r wrappedReconciler[T]) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	typ := reflect.ValueOf(*new(T)).Type().Elem()
	obj := reflect.New(typ).Interface().(T)

	err := r.client.Get(ctx, req.NamespacedName, obj)
	err = client.IgnoreNotFound(err)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := r.next.Reconcile(ctx, obj)
	if err != nil {
		r.recorder.Event(obj, corev1.EventTypeWarning, "Error", "Skiperator has encountered a problem")
	}

	return res, err
}
