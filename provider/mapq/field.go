package mapq

import "github.com/savaki/graphql"

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
