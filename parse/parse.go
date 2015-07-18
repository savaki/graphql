package parse

func Parse(q string) (*Node, error) {
	l := lex("graph", q)
	iter := newIterator(l)

	parse(iter)
	if iter.err != nil {
		return nil, iter.err
	}

	return &Node{Query: iter.selector}, nil
}

func parse(iter *iterator) {
	var fn parseFn = parseRoot
	for fn != nil {
		fn = fn(iter)
	}
}

func parseRoot(iter *iterator) parseFn {
	item := iter.peek()
	switch item.typ {
	case itemQuery:
		iter.next()
		return parseSelector

	default:
		return iter.errorf("unexpected element after query => %s", item.typ)
	}
}

func parseSelector(iter *iterator) parseFn {
	item := iter.peek()
	switch {
	case item.typ == itemName && iter.peek1().typ == itemColon:
		alias := iter.next() // alias
		iter.next()          // colon
		name := iter.next()  // name

		iter.addAlias(alias.val, name.val)
		return parseField

	case item.typ == itemName:
		iter.next()
		iter.addField(item.val)
		return parseField

	case item.typ == itemRightCurly:
		iter.next()
		iter.popSelector()
		return parseSelector

	case item.typ == itemEOF:
		return nil

	default:
		return iter.errorf("unexpected element after query => %s", item.typ)
	}
}

func parseField(iter *iterator) parseFn {
	item := iter.peek()
	switch {
	case item.typ == itemLeftParen:
		iter.next()
		return parseFieldArg

	case item.typ == itemLeftCurly:
		iter.next()
		s := iter.addSelector()
		iter.pushSelector(s)
		return parseSelector

	case item.typ == itemDot:
		iter.next()
		return parseOperation

	case item.typ == itemRightCurly:
		iter.next()
		iter.popSelector()
		return parseSelector

	default:
		return iter.errorf("unexpected element after name => %s", item.typ)
	}
}

func parseOperation(iter *iterator) parseFn {
	item := iter.peek()
	switch {
	case item.typ == itemName && iter.peek1().typ == itemLeftParen:
		iter.next() // name
		iter.next() // left paren
		iter.addOperation(item.val)
		return parseOperationArg

	default:
		return iter.errorf("unexpected operation after dot => %s", item.typ)
	}
	return nil
}

func parseOperationArg(iter *iterator) parseFn {
	item := iter.peek()
	switch {
	case item.typ == itemName && iter.peek1().typ == itemColon:
		name := iter.next()  // name
		iter.next()          // colon
		value := iter.next() // value

		iter.addOperationArg(name.val, value.val)
		return parseOperationArg

	case item.typ == itemRightParen:
		iter.next()
		return parseField

	case item.typ == itemNumber:
		iter.next()
		iter.addOperationArg("", item.val)
		return parseOperationArg

	default:
		return iter.errorf("unexpected operation argument element => %s", item.typ)
	}

}

func parseFieldArg(iter *iterator) parseFn {
	item := iter.peek()
	switch {
	case item.typ == itemName && iter.peek1().typ == itemColon:
		name := iter.next()  // name
		iter.next()          // colon
		value := iter.next() // value

		iter.addFieldArg(name.val, value.val)
		return parseFieldArg

	case item.typ == itemRightParen:
		iter.next()
		return parseField

	default:
		return iter.errorf("unexpected field argument element => %s", item.typ)
	}
}
