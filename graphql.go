package gographql

type Value interface {
}

type Arg struct {
	Name  string
	Value Value
}

type Selection interface {
	Fetch(string, ...Arg) (Field, error)
}

type Field interface {
	Selection() (Selection, error)
	Apply(string, ...Arg) (Field, error)
	Value() (Value, error)
}

type Store interface {
	Query(string, ...Arg) (Field, error)
	Mutate(string, ...Arg) (Field, error)
}
