package util

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type MatchesPredicate[T client.Object] func(T) bool

func (m MatchesPredicate[T]) Create(evt event.CreateEvent) bool {
	return m(evt.Object.(T))
}

func (m MatchesPredicate[T]) Delete(evt event.DeleteEvent) bool {
	return m(evt.Object.(T))
}

func (m MatchesPredicate[T]) Update(evt event.UpdateEvent) bool {
	return m(evt.ObjectNew.(T))
}

func (m MatchesPredicate[T]) Generic(evt event.GenericEvent) bool {
	return m(evt.Object.(T))
}
