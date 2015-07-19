package parse

// --[ Arg ]----------------------------------------------------------

type Arg struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
}

// --[ Filter ]-------------------------------------------------------

type Filter struct {
	Name string `json:"name,omitempy"`
	Args []*Arg `json:"args,omitempty"`
}

func (op *Filter) addArg(name, value string) *Arg {
	arg := &Arg{
		Name:  name,
		Value: value,
	}
	op.Args = append(op.Args, arg)
	return arg
}

// --[ Field ]--------------------------------------------------------

type Field struct {
	Alias      string    `json:"alias,omitempty"`
	Name       string    `json:"name"`
	Args       []*Arg    `json:"args,omitempty"`
	Selector   *Selector `json:"selector,omitempty"`
	Operations []*Filter `json:"operations,omitempty"`
}

func (f*Field) IsScalar() bool {
	return f.Selector == nil || len(f.Selector.Fields) == 0
}

func (f *Field) addArg(name, value string) *Arg {
	arg := &Arg{
		Name:  name,
		Value: value,
	}
	f.Args = append(f.Args, arg)
	return arg
}

func (f *Field) addSelector() *Selector {
	f.Selector = &Selector{}
	return f.Selector
}

func (f *Field) addFilter(name string) *Filter {
	op := &Filter{Name: name}
	f.Operations = append(f.Operations, op)
	return op
}

func newField(alias, name string) *Field {
	return &Field{
		Alias: alias,
		Name:  name,
	}
}

// --[ Selector ]-----------------------------------------------------

type Selector struct {
	Fields []*Field `json:"fields,omitempty"`
}

func (s *Selector) addAlias(alias, name string) *Field {
	field := &Field{
		Alias: alias,
		Name:  name,
	}
	s.Fields = append(s.Fields, field)
	return field
}

func (s *Selector) addField(name string) *Field {
	return s.addAlias("", name)
}

// --[ Operations ]---------------------------------------------------

//go:generate go get github.com/campoy/jsonenums
//go:generate jsonenums -type=OperationType
type OperationType int

const (
	OpUnknown OperationType = iota
	OpQuery
	OpMutation
)

type Operation struct {
	Type  OperationType `json:"type"`
	Field *Field        `json:"field"`
}

func newOperation(opType OperationType, alias, name string) *Operation {
	return &Operation{
		Type:  opType,
		Field: newField(alias, name),
	}
}

// --[ Document ]-----------------------------------------------------

type Document struct {
	Operations []*Operation `json:"operations"`
}

func (n *Document) addOperation(opType OperationType, name string) *Operation {
	op := &Operation{
		Type:  opType,
		Field: newField("", name),
	}
	n.Operations = append(n.Operations, op)
	return op
}

type parseFn func(iter *iterator) parseFn
