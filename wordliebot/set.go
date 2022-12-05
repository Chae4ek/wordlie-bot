package wordliebot

type void struct{}

var memberOfSet void

type Set[T comparable] struct {
	elements map[T]void
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{elements: make(map[T]void)}
}

func (set *Set[T]) Add(element T) {
	set.elements[element] = memberOfSet
}

func (set *Set[T]) Delete(element T) {
	delete(set.elements, element)
}

func (set *Set[T]) Contains(element T) bool {
	_, exists := set.elements[element]
	return exists
}
