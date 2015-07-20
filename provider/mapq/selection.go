package mapops

import "github.com/savaki/graphql"

type selection struct {
	data map[string]interface{}
}

func New(data map[string]interface{}) gographql.Store {
	return &selection{
		data: data,
	}
}

func (s *selection) Fetch(name string, args ...gographql.Arg) (gographql.Field, error) {
	v, ok := s.data[name]
	if !ok {
		return nil, errFieldNotFound
	}
	return &field{value: v}, nil
}

func (s *selection) Query(name string, args ...gographql.Arg) (gographql.Field, error) {
	return s.Fetch(name, args...)
}

func (s *selection) Mutate(name string, args ...gographql.Arg) (gographql.Field, error) {
	return nil, errNotImplemented
}
