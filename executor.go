package gographql

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/savaki/graphql/parse"
)

var (
	ErrUnknownQuery = errors.New("unknown query operation")
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
	args := make([]Arg, len(qOp.Field.Args))
	for index, arg := range qOp.Field.Args {
		args[index] = Arg{
			Name:  arg.Name,
			Value: arg.Value,
		}
	}
	field, err := store.Query(qOp.Field.Name, args...)
	if err != nil {
		return ErrUnknownQuery
	}

	return writeField(w, field, qOp.Field)
}

func writeField(w io.Writer, field Field, qField *parse.Field) error {
	label := qField.Name
	if qField.Alias != "" {
		label = qField.Alias
	}

	io.WriteString(w, `"`)
	io.WriteString(w, label)
	io.WriteString(w, `":`)

	if qField.IsScalar() {
		return writeValue(w, field, qField)
	} else {
		selection, err := field.Selection()
		if err != nil {
			return err
		}
		return writeSelection(w, selection, qField.Selector)
	}
}

func writeValue(w io.Writer, field Field, qField *parse.Field) error {
	v, err := field.Value()
	if err != nil {
		return err
	}

	return json.NewEncoder(w).Encode(v)
}

func writeSelection(w io.Writer, selection Selection, qSelector *parse.Selector) error {
	io.WriteString(w, "{")
	for _, qField := range qSelector.Fields {
		field, err := selection.Fetch(qField.Name)
		if err != nil {
			return err
		}
		writeField(w, field, qField)
	}
	io.WriteString(w, "}")

	return nil
}
