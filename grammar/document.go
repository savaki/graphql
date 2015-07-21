package grammar

// --[ Arg ]----------------------------------------------------------

type Arg struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
}

// --[ Filter ]-------------------------------------------------------

type Filter struct {
	Name string `json:"name,omitempty"`
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
	Alias      string     `json:"alias,omitempty"`
	Name       string     `json:"name,omitempty"`
	Args       []*Arg     `json:"args,omitempty"`
	Selection  *Selection `json:"selector,omitempty"`
	Operations []*Filter  `json:"operations,omitempty"`
}

func (f *Field) Key() string {
	if f.Alias == "" {
		return f.Name
	} else {
		return f.Alias
	}
}

func (f *Field) IsScalar() bool {
	return f.Selection == nil || len(f.Selection.Fields) == 0
}

func (f *Field) addArg(name, value string) *Arg {
	arg := &Arg{
		Name:  name,
		Value: value,
	}
	f.Args = append(f.Args, arg)
	return arg
}

func (f *Field) addSelection() *Selection {
	f.Selection = &Selection{}
	return f.Selection
}

func newField(alias, name string) *Field {
	return &Field{
		Alias: alias,
		Name:  name,
	}
}

// --[ Selector ]-----------------------------------------------------

type Selection struct {
	Fields []*Field `json:"fields,omitempty"`
}

func (s *Selection) addAlias(alias, name string) *Field {
	field := &Field{
		Alias: alias,
		Name:  name,
	}
	s.Fields = append(s.Fields, field)
	return field
}

func (s *Selection) addField(name string) *Field {
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
	Field *Field        `json:"field,omitempty"`
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

func (d *Document) HasDefaultQueryOnly() bool {
	return len(d.Operations) == 1 && d.Operations[0].Field.Name == ""
}

func (d *Document) addOperation(opType OperationType, name string) *Operation {
	op := &Operation{
		Type:  opType,
		Field: newField("", name),
	}
	d.Operations = append(d.Operations, op)
	return op
}

type parseFn func(iter *iterator) parseFn
