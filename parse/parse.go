package parse

func Parse(q string) (*Document, error) {
	l := lex("graph", q)
	iter := newIterator(l)

	parse(iter)
	if iter.err != nil {
		return nil, iter.err
	}

	return &Document{Operations: iter.operations}, nil
}

func parse(iter *iterator) {
	var fn parseFn = parseRoot
	for fn != nil {
		fn = fn(iter)
	}
}

func parseRoot(iter *iterator) parseFn {
	item := iter.peek()
	switch {
	case item.typ == itemLeftCurly:
		iter.next()
		iter.addQuery("", "")
		iter.addSelection()
		return parseSelector

	case item.typ == itemQuery:
		iter.next()
		return parseQuery

	default:
		return iter.errorf("unexpected element in root => %s", item.typ)
	}
}

func parseQuery(iter *iterator) parseFn {
	item := iter.peek()
	item1 := iter.peek1()
	item2 := iter.peek2()

	switch {
	case item.typ == itemName && item1.typ == itemColon && item2.typ == itemName:
		alias := iter.next() // alias
		iter.next()          // colon
		name := iter.next()  // name

		iter.addQuery(alias.val, name.val)
		return parseField

	case item.typ == itemName:
		name := iter.next() // name

		iter.addQuery("", name.val)
		return parseField

	default:
		return iter.errorf("unexpected element after query => %s", item.typ)
	}
}

func parseSelector(iter *iterator) parseFn {
	item := iter.peek()
	item1 := iter.peek1()
	item2 := iter.peek2()

	switch {
	case item.typ == itemName && item1.typ == itemColon && item2.typ == itemName:
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
	item1 := iter.peek1()
	item2 := iter.peek2()

	switch {
	case item.typ == itemLeftParen:
		iter.next()
		return parseFieldArg

	case item.typ == itemLeftCurly:
		iter.next()
		s := iter.addSelection()
		iter.pushSelector(s)
		return parseSelector

	case item.typ == itemRightCurly:
		iter.next()
		iter.popSelector()
		return parseSelector

	case item.typ == itemName && item1.typ == itemColon && item2.typ == itemName:
		alias := iter.next() // alias
		iter.next()          // colon
		name := iter.next()  // name

		iter.addAlias(alias.val, name.val)
		return parseField

	case item.typ == itemName:
		iter.next()
		iter.addField(item.val)
		return parseField

	default:
		iter.dumpTokens()
		return iter.errorf("unexpected element after name => %s", item.typ)
	}
}

func parseFieldArg(iter *iterator) parseFn {
	item := iter.peek()
	item1 := iter.peek1()
	item2 := iter.peek2()

	switch {
	case item.typ == itemName && item1.typ == itemColon && isValue(item2):
		name := iter.next()  // name
		iter.next()          // colon
		value := iter.next() // value

		iter.addFieldArg(name.val, value.val)
		return parseFieldArg

	case isValue(item):
		value := iter.next() // value

		iter.addFieldArg("", value.val)
		return parseFieldArg

	case item.typ == itemRightParen:
		iter.next()
		return parseField

	default:
		return iter.errorf("unexpected field argument element => %s", item.typ)
	}
}

func isValue(item item) bool {
	return item.typ == itemInt || item.typ == itemString
}
