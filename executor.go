package graphql

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/savaki/graphql/parse"
)

type Executor struct {
	Store Store
}

func New(store Store) Executor {
	return Executor{
		Store: store,
	}
}

func (e Executor) Handle(query string, w io.Writer) error {
	doc, err := parse.Parse(string(query))
	if err != nil {
		return err
	}

	return writeDocument(w, e.Store, doc)
}

func writeDocument(w io.Writer, store Store, doc *parse.Document) error {
	io.WriteString(w, "{")
	for index, ops := range doc.Operations {
		err := writeOperation(w, store, ops)
		if err != nil {
			return err
		}
		if index < len(doc.Operations)-1 {
			io.WriteString(w, ",")
		}
	}
	io.WriteString(w, "}")

	return nil
}

func writeOperation(w io.Writer, store Store, qOp *parse.Operation) error {
	io.WriteString(w, `"`)
	if qOp.Field.Alias == "" {
		io.WriteString(w, qOp.Field.Name)
	} else {
		io.WriteString(w, qOp.Field.Alias)
	}
	io.WriteString(w, `":`)

	args := make([]Arg, len(qOp.Field.Args))
	for index, arg := range qOp.Field.Args {
		args[index] = Arg{
			Name:  arg.Name,
			Value: arg.Value,
		}
	}
	ctx := &Context{Name: qOp.Field.Name, Args: args}
	selection, err := store.Query(ctx)
	if err != nil {
		return ErrUnknownQuery
	}

	return writeSelection(w, selection, qOp.Field.Selection)
}

func writeSelection(w io.Writer, selection Selection, qSelector *parse.Selector) error {
	io.WriteString(w, "{")
	for index, qField := range qSelector.Fields {
		ctx := &Context{Name: qField.Name}
		field, err := selection.Fetch(ctx)
		if err != nil {
			return err
		}
		writeField(w, field, qField)

		if index < len(qSelector.Fields)-1 {
			io.WriteString(w, ",")
		}
	}
	io.WriteString(w, "}")

	return nil
}

func writeField(w io.Writer, field Field, qField *parse.Field) error {
	io.WriteString(w, `"`)
	io.WriteString(w, qField.Key())
	io.WriteString(w, `":`)

	if qField.IsScalar() {
		return writeValue(w, field, qField)

	} else {
		selection, err := field.Selection()
		if err != nil {
			return err
		}
		return writeSelection(w, selection, qField.Selection)
	}
}

func writeValue(w io.Writer, field Field, qField *parse.Field) error {
	v, err := field.Value()
	if err != nil {
		return err
	}

	if v == nil {
		_, err = io.WriteString(w, "nil")
		return err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
