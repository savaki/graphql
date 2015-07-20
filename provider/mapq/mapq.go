package mapq

import (
	"errors"

	"github.com/savaki/graphql"
)

var (
	errNotImplemented = errors.New("not implemented")
	errFieldNotFound  = errors.New("field not found")
)

// --[ Field ]------------------------------------------------------------

type field struct {
	value interface{}
}

func (f *field) Value() (graphql.Value, error) {
	return f.value, nil
}

func (f *field) Selection() (graphql.Selection, error) {
	switch v := f.value.(type) {
	case map[string]interface{}:
		return &selection{data: v}, nil
	}

	return nil, errNotImplemented
}

// --[ Store / Selection ]------------------------------------------------

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
