package parse

// --[ Arg ]----------------------------------------------------------

type Arg struct {
	Name  string `json:"name,omitempy"`
	Value string `json:"value"`
}

// --[ Operations ]---------------------------------------------------

type Operation struct {
	Name string `json:"name,omitempy"`
	Args []*Arg `json:"args,omitempty"`
}

func (op *Operation) addArg(name, value string) *Arg {
	arg := &Arg{
		Name:  name,
		Value: value,
	}
	op.Args = append(op.Args, arg)
	return arg
}

// --[ Field ]--------------------------------------------------------

type Field struct {
	Alias      string       `json:"alias,omitempty"`
	Name       string       `json:"name"`
	Args       []*Arg       `json:"args,omitempty"`
	Selector   *Selector    `json:"selector,omitempty"`
	Operations []*Operation `json:"operations,omitempty"`
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

func (f *Field) addOperation(name string) *Operation {
	op := &Operation{Name: name}
	f.Operations = append(f.Operations, op)
	return op
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

type Node struct {
	Query *Selector `json:"query"`
}

type parseFn func(iter *iterator) parseFn
