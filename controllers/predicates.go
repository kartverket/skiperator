package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type matchesPredicate[T client.Object] func(T) bool

func (m matchesPredicate[T]) Create(evt event.CreateEvent) bool {
	return m(evt.Object.(T))
}

func (m matchesPredicate[T]) Delete(evt event.DeleteEvent) bool {
	return m(evt.Object.(T))
}

func (m matchesPredicate[T]) Update(evt event.UpdateEvent) bool {
	return m(evt.ObjectNew.(T))
}

func (m matchesPredicate[T]) Generic(evt event.GenericEvent) bool {
	return m(evt.Object.(T))
}
