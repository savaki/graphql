package restq

import (
	"github.com/docker/machine/drivers/vmwarevsphere/errors"
	"github.com/savaki/graphql"
)

type field struct {
	selection graphql.Selection
}

func (f field) Value() (graphql.Value, error) {
	return nil, errors.New("illegal call to Value; not a scalar value")
}

func (f field) Selection() (graphql.Selection, error) {
	return f.selection, nil
}
