package pubsub

import "sync"

type PubSub struct {
	mu   sync.Mutex
	subs map[string][]chan string
}

func NewPubSub() *PubSub {
	return &PubSub{
		mu:   sync.Mutex{},
		subs: make(map[string][]chan string),
	}
}

func (ps *PubSub) Subscribe(topic string) <-chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan string)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *PubSub) Publish(topic string, data string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, ch := range ps.subs[topic] {
		ch <- data
	}
}
