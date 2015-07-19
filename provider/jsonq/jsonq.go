package jsonq

import (
	"encoding/json"

	"github.com/savaki/gographql"
)

// --[ Field ]------------------------------------------------------------

type Field struct {
	data json.RawMessage
}

func (f Field) Apply(name string, args ...gographql.Arg) (gographql.Field, error) {
	return nil, gographql.ErrNotImplemented
}

func (f Field) Selection() (gographql.Selection, error) {
	s, err := New(f.data)
	return s, err
}

func (f Field) Value() (gographql.Value, error) {
	switch f.data[0] {
	case '"':
		var s string
		err := json.Unmarshal(f.data, &s)
		return s, err

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		var i int
		err := json.Unmarshal(f.data, &i)
		return i, err

	case '{', '[':
		return nil, gographql.ErrNotAScalar
	}

	return nil, nil
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

func (s Store) Fetch(name string, args ...gographql.Arg) (gographql.Field, error) {
	v, ok := s.props[name]
	if !ok {
		return nil, gographql.ErrFieldNotFound
	}

	return Field{data: v}, nil
}

func (s Store) Query(name string, args ...gographql.Arg) (gographql.Field, error) {
	return s.Fetch(name)
}

func (s Store) Mutate(name string, args ...gographql.Arg) (gographql.Field, error) {
	return nil, gographql.ErrNotImplemented
}
