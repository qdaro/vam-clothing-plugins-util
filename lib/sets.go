package lib

import "regexp"

// Generic set
type Set[T comparable] struct {
	Data map[T]bool
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		Data: make(map[T]bool),
	}
}

func NewSetFrom[T comparable](source []T) *Set[T] {
	set := &Set[T]{
		Data: make(map[T]bool),
	}
	set.Add(source...)
	return set
}

func (s *Set[T]) Add(elements ...T) {
	for _, element := range elements {
		s.Data[element] = true
	}
}

func (s *Set[T]) Append(elements []T) {
	for _, element := range elements {
		s.Data[element] = true
	}
}

func (s *Set[T]) Contains(element T) bool {
	_, exists := s.Data[element]
	return exists
}

func (s *Set[T]) Remove(element T) {
	delete(s.Data, element)
}

func (s *Set[T]) Elements() []T {
	elements := make([]T, 0, len(s.Data))
	for key := range s.Data {
		elements = append(elements, key)
	}
	return elements
}

func (s *Set[T]) Size() int {
	return len(s.Data)
}

func (s *Set[T]) Clear() {
	s.Data = make(map[T]bool)
}

// Set of string paths with relevant methods
type PathSet struct {
	// Embed the Set struct
	Set[string]
}

func NewPathSet() *PathSet {
	return &PathSet{
		Set: *NewSet[string](),
	}
}

func (s *PathSet) GetFirstMatching(exp *regexp.Regexp) (path string, found bool) {
	for p := range s.Data {
		if exp.MatchString(p) {
			return p, true
		}
	}

	return "", false
}

func (s *PathSet) GetAllMatching(exp *regexp.Regexp) (paths []string) {
	var matched []string

	for p := range s.Data {
		if exp.MatchString(p) {
			matched = append(matched, p)
		}
	}

	return matched
}
