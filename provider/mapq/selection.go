package mapops

import "github.com/savaki/graphql"

type selection struct {
	data map[string]interface{}
}

func New(data map[string]interface{}) graphql.Store {
	return &selection{
		data: data,
	}
}

func (s *selection) Fetch(c *graphql.Context) (graphql.Field, error) {
	v, ok := s.data[c.Name]
	if !ok {
		return nil, errFieldNotFound
	}
	return &field{value: v}, nil
}

func (s *selection) Query(c *graphql.Context) (graphql.Selection, error) {
	f, err := s.Fetch(c)
	if err != nil {
		return nil, err
	}

	v, err := f.Selection()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *selection) Mutate(c *graphql.Context) (graphql.Selection, error) {
	return nil, errNotImplemented
}
