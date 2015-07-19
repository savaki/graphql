package parse

import (
	"errors"
	"fmt"
)

type iterator struct {
	l          *lexer
	err        error
	operations []*Operation
	operation  *Operation
	selectors  []*Selector
	selector   *Selector
	field      *Field
	filter     *Filter

	tokens    [16]item
	tokensPos Pos
	tokensEOF bool
}

func newIterator(l *lexer) *iterator {
	iter := &iterator{
		l:         l,
		selectors: []*Selector{},
	}
	iter.pushSelector(&Selector{})

	for i := 0; i < len(iter.tokens); i++ {
		item := l.nextItem()
		iter.tokens[i] = item
		if item.typ == itemEOF {
			iter.tokensEOF = true
			break
		}
	}

	return iter
}

func (iter *iterator) next() item {
	if iter.tokensPos == eof {
		panic("invalid call to next; past end of stream")
	}

	// retrieve the next value
	item := iter.tokens[iter.tokensPos]

	// fetch next element unless we've come to the end
	if !iter.tokensEOF {
		item := iter.l.nextItem()
		iter.tokens[iter.tokensPos] = item
		iter.tokensEOF = (item.typ == itemEOF)
	}

	// advance the pointer
	iter.tokensPos = (iter.tokensPos + 1) % Pos(len(iter.tokens))

	return item
}

func (iter *iterator) peek() item {
	return iter.peekN(0)
}

func (iter *iterator) peek1() item {
	return iter.peekN(1)
}

func (iter *iterator) peek2() item {
	return iter.peekN(2)
}

// peekN peeks N elements into the future
func (iter *iterator) peekN(n Pos) item {
	if int(n) > len(iter.tokens) {
		panic(fmt.Sprintf("illegal attempt to peek too far into the future, max: %v", len(iter.tokens)))
	}

	pos := (iter.tokensPos + n) % Pos(len(iter.tokens))
	return iter.tokens[pos]
}

func (iter *iterator) errorf(format string, args ...interface{}) parseFn {
	iter.err = errors.New(fmt.Sprintf(format, args...))
	return nil
}

func (iter *iterator) addQuery(alias, name string) *Operation {
	iter.operation = newOperation(OpQuery, alias, name)
	iter.operations = append(iter.operations, iter.operation)
	iter.field = iter.operation.Field
	return iter.operation
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

func (iter *iterator) addFilterArg(name, value string) {
	if iter.filter != nil {
		iter.filter.addArg(name, value)
	}
}

func (iter *iterator) addSelector() *Selector {
	if iter.field != nil {
		return iter.field.addSelector()
	} else {
		return &Selector{}
	}
}

func (iter *iterator) addFilter(name string) *Filter {
	iter.filter = iter.field.addFilter(name)
	return iter.filter
}

func (iter *iterator) pushSelector(s *Selector) {
	iter.selectors = append(iter.selectors, s)
	iter.selector = s
	iter.field = nil
}

func (iter *iterator) popSelector() *Selector {
	length := len(iter.selectors)
	s := iter.selectors[length-1]
	iter.selectors = iter.selectors[0 : length-1]

	iter.selector = s
	iter.field = nil

	return s
}
