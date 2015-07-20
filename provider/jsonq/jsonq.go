package jsonq

import (
	"encoding/json"

	"github.com/savaki/graphql"
)

// --[ Field ]------------------------------------------------------------

type Field struct {
	data json.RawMessage
}

func (f Field) Apply(name string, args ...graphql.Arg) (graphql.Field, error) {
	return nil, graphql.ErrNotImplemented
}

func (f Field) Selection() (graphql.Selection, error) {
	s, err := New(f.data)
	return s, err
}

func (f Field) Value() (graphql.Value, error) {
	switch f.data[0] {
	case '"':
		var s string
		err := json.Unmarshal(f.data, &s)
		return s, err

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		var i float32
		err := json.Unmarshal(f.data, &i)
		return i, err

	case '{', '[':
		return nil, graphql.ErrNotAScalar

	default:
		return nil, graphql.ErrNotAScalar
	}
}

// --[ Store ]------------------------------------------------------------

type Store struct {
	props map[string]json.RawMessage
}

func New(data []byte) (Store, error) {
	v := map[string]json.RawMessage{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return Store{}, err
	}

	return Store{props: v}, nil
}

func (s Store) Fetch(c *graphql.Context) (graphql.Field, error) {
	v, ok := s.props[c.Name]
	if !ok {
		return nil, graphql.ErrFieldNotFound
	}

	return Field{data: v}, nil
}

func (s Store) Query(c *graphql.Context) (graphql.Field, error) {
	return s.Fetch(c)
}

func (s Store) Mutate(c *graphql.Context) (graphql.Field, error) {
	return nil, graphql.ErrNotImplemented
}
