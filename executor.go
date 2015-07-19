package gographql

import "github.com/savaki/gographql/parse"

type Executor struct {
	stores map[string]Store
}

func (f *Executor) Handle(query string) ([]byte, error) {
	_, err := parse.Parse(query)
	if err != nil {
		return nil, err
	}

//	for _, field := range n.Operations.Fields {
//		if _, ok := f.stores[field.Name]; ok {
////			store.Query()
//		}
//	}

	return nil, nil
}
