package graphql

import "errors"

var (
	ErrFieldNotFound  = errors.New("field not found")
	ErrNotImplemented = errors.New("feature not implemented")
	ErrNotAScalar     = errors.New("invalid attempt to treat non-scalar as scalar")
	ErrUnknownQuery   = errors.New("unknown query operation")
)

// --[ Value ]--------------------------------------------------------

type Value interface {
}

// --[ Arg ]----------------------------------------------------------

type Arg struct {
	Name  string
	Value Value
}

// --[ Filter ]-------------------------------------------------------

type Filter struct {
	name string
	args []Arg
}

func (f *Filter) Name() string {
	return f.name
}

func (f *Filter) Args() []Arg {
	return f.args
}

// --[ Context ]------------------------------------------------------

type Context struct {
	Name    string
	Args    []Arg
	Filters []Filter
}

// --[ Selection ]----------------------------------------------------

type Query interface {
	Query(*Context) (Field, error)
}

type Selection interface {
	Query
}

type Field interface {
	Selection() (Selection, error)
	Value() (Value, error)
}

type Store interface {
	Query
	Mutate(*Context) (Field, error)
}
