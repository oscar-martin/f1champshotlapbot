package pubsub

import (
	"sync"
)

type PubSub[T any] struct {
	mu   sync.Mutex
	subs map[string][]chan T
}

func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{
		subs: make(map[string][]chan T),
	}
}

func (ps *PubSub[T]) Subscribe(topic string) <-chan T {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan T)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *PubSub[T]) Publish(topic string, data T) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, ch := range ps.subs[topic] {
		// fmt.Printf("Publishing to %s\n", topic)
		ch <- data
		// fmt.Printf("Published to %s\n", topic)
	}
}
