package mapops

import "github.com/savaki/gographql"

type field struct {
	value interface{}
}

func (f *field) Apply(name string, args ...gographql.Arg) (gographql.Field, error) {
	return nil, errFieldNotFound
}

func (f *field) Value() (gographql.Value, error) {
	return f.value, nil
}

func (f *field) Selection() (gographql.Selection, error) {
	switch v := f.value.(type) {
	case map[string]interface{}:
		return &selection{data: v}, nil
	}

	return nil, errNotImplemented
}
