package parse

import (
	"errors"
	"fmt"
)

type iterator struct {
	l         *lexer
	token     [3]item
	empty     bool
	err       error
	stack     []*Selector
	selector  *Selector
	field     *Field
	operation *Operation
}

func newIterator(l *lexer) *iterator {
	iter := &iterator{
		l:     l,
		stack: []*Selector{},
	}
	iter.pushSelector(&Selector{})

	for i := 0; i < len(iter.token); i++ {
		item := l.nextItem()
		iter.token[i] = item
		if item.typ == itemEOF {
			iter.empty = true
			break
		}
	}

	return iter
}

func (iter *iterator) next() item {
	item := iter.token[0]
	iter.token[0] = iter.token[1]
	iter.token[1] = iter.token[2]

	if !iter.empty {
		next := iter.l.nextItem()
		iter.token[2] = next
		if next.typ == itemEOF {
			iter.empty = true
		}
	}

	return item
}

func (iter *iterator) peek() item {
	return iter.token[0]
}

// peek1 peeks one element into the future
func (iter *iterator) peek1() item {
	return iter.token[1]
}

func (iter *iterator) errorf(format string, args ...interface{}) parseFn {
	iter.err = errors.New(fmt.Sprintf(format, args...))
	return nil
}

func (iter *iterator) addAlias(alias, name string) {
	iter.field = iter.selector.addAlias(alias, name)
}

func (iter *iterator) addField(name string) {
	iter.field = iter.selector.addField(name)
}

func (iter *iterator) addFieldArg(name, value string) {
	if iter.field != nil {
		iter.field.addArg(name, value)
	}
}

func (iter *iterator) addOperationArg(name, value string) {
	if iter.operation != nil {
		iter.operation.addArg(name, value)
	}
}

func (iter *iterator) addSelector() *Selector {
	if iter.field != nil {
		return iter.field.addSelector()
	} else {
		return &Selector{}
	}
}

func (iter *iterator) addOperation(name string) *Operation {
	iter.operation = iter.field.addOperation(name)
	return iter.operation
}

func (iter *iterator) pushSelector(s *Selector) {
	iter.stack = append(iter.stack, s)
	iter.selector = s
	iter.field = nil
}

func (iter *iterator) popSelector() *Selector {
	length := len(iter.stack)
	fmt.Println("stack depth =>", length)
	s := iter.stack[length-1]
	iter.stack = iter.stack[0 : length-1]

	iter.selector = s
	iter.field = nil

	return s
}
