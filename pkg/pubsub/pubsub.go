package pubsub

type PubSub struct {
	subs map[string][]chan string
}

func NewPubSub() *PubSub {
	return &PubSub{
		subs: make(map[string][]chan string),
	}
}

func (ps *PubSub) Subscribe(topic string) <-chan string {
	ch := make(chan string)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *PubSub) Publish(topic string, data string) {
	for _, ch := range ps.subs[topic] {
		ch <- data
	}
}
