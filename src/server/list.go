package brev

type node[T any] struct {
	value T

	next *node[T]
	prev *node[T]
}

type list[T any] struct {
	head *node[T]
	tail *node[T]
}

func (l *list[T]) Head() *node[T] {
	return l.head
}

func (l *list[T]) Tail() *node[T] {
	return l.tail
}

func (l *list[T]) Append(value T) {
	n := &node[T]{
		value: value,
		prev:  l.tail,
	}

	if l.tail != nil {
		l.tail.next = n
		l.tail = n
	}
}
