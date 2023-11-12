package queues

type Queue[T any] []T

func NewQueue[T any]() *Queue[T] {
	q := Queue[T]{}
	return &q
}

func (q *Queue[T]) Push(x T) {
	*q = append(*q, x)
}

func (q *Queue[T]) Peek() T {
	return (*q)[0]
}

func (q *Queue[T]) Pop() T {
	x := (*q)[0]
	*q = (*q)[1:]
	return x
}

func (q *Queue[T]) IsEmpty() bool {
	return len(*q) == 0
}
