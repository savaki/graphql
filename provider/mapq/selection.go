package mapq

import "github.com/savaki/graphql"

type selection struct {
	data map[string]interface{}
}

func New(data map[string]interface{}) graphql.Store {
	return &selection{
		data: data,
	}
}

func (s *selection) Query(c *graphql.Context) (graphql.Field, error) {
	v, ok := s.data[c.Name]
	if !ok {
		return nil, errFieldNotFound
	}
	return &field{value: v}, nil
}

func (s *selection) Mutate(c *graphql.Context) (graphql.Field, error) {
	return nil, errNotImplemented
}
