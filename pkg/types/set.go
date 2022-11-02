package types

import (
	"fmt"
	"strings"
)

type Set[T comparable] struct {
	internal map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	s := new(Set[T])
	s.internal = map[T]struct{}{}
	return s
}

func (s *Set[T]) Add(e T) {
	s.internal[e] = struct{}{}
}

func (s *Set[T]) Remove(e T) {
	delete(s.internal, e)
}

func (s *Set[T]) Has(e T) bool {
	_, has := s.internal[e]
	return has
}

func (s *Set[T]) String(e T) string {
	b := strings.Builder{}
	b.WriteString("Set [ ")
	for e, _ := range s.internal {
		b.WriteString(fmt.Sprintf(" %v ", e))
	}
	b.WriteString(" ] ")
	return b.String()
}
