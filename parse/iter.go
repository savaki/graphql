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
	selectors  []*Selection
	selection  *Selection
	field      *Field
	filter     *Filter

	tokens    [16]item
	tokensPos Pos
	tokensEOF bool
}

func newIterator(l *lexer) *iterator {
	iter := &iterator{
		l:         l,
		selectors: []*Selection{},
	}
	iter.pushSelector(&Selection{})

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
	iter.field = iter.selection.addAlias(alias, name)
}

func (iter *iterator) addField(name string) {
	iter.field = iter.selection.addField(name)
}

func (iter *iterator) addFieldArg(name, value string) {
	if iter.field != nil {
		iter.field.addArg(name, value)
	}
}

func (iter *iterator) addSelection() *Selection {
	iter.selection = iter.field.addSelection()
	return iter.selection
}

func (iter *iterator) pushSelector(s *Selection) {
	iter.selectors = append(iter.selectors, s)
	iter.selection = s
	iter.field = nil
}

func (iter *iterator) popSelector() *Selection {
	length := len(iter.selectors)
	s := iter.selectors[length-1]
	iter.selectors = iter.selectors[0 : length-1]

	iter.selection = s
	iter.field = nil

	return s
}
