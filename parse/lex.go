// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

//go:generate stringer -type=itemType
// itemType identifies the type of lex items.
type itemType int

const (
	itemError        itemType = iota // error occurred; value is text of error
	itemBool                         // boolean constant
	itemChar                         // printable ASCII character; grab bag for comma etc.
	itemCharConstant                 // character constant
	itemComplex                      // complex constant (1+2i); imaginary is just a number
	itemEOF
	itemField         // alphanumeric identifier starting with '.'
	itemIdentifier    // alphanumeric identifier not starting with '.'
	itemQueryKeyword // query keyword
	itemQueryName     // the name for a query
	itemQueryField    // the field within the query
	itemQueryBegin    // marks start of query block
	itemQueryEnd      // marks end of query block
	itemParamBegin    // '(' inside action
	itemParamEnd      // ')' inside action
	itemNumber        // simple number, including imaginary
	itemRawString     // raw quoted string (includes quotes)
	itemRightDelim    // right action delimiter
	itemSpace         // run of spaces separating arguments
	itemString        // quoted string (includes quotes)
	itemText          // plain text
	itemColon         // the : separating param name from param value
	itemParamName     // the parameter name
	itemSelectorBegin // marks the start of a selector block
	itemSelectorEnd   // marks the end of a selector block
	itemSelector      // marks the field we want returned

	// ONLY KEYWORDS BELOW THIS POINT
	itemKeyword      // used only to delimit the keywords
	itemDot          // the cursor, spelled '.'
	itemNil          // the untyped nil constant, easiest to treat as a keyword
)

var keywords = map[itemType]string{
	itemQueryKeyword: "query",
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) int {
	count := 0
	for strings.IndexRune(valid, l.next()) >= 0 {
		count++
	}
	l.backup()
	return count
}

// acceptFn consumes a run of runes matched by the valid validFn; returns length of run
func (l *lexer) acceptFn(validFn func(rune) bool) int {
	count := 0
	for validFn(l.next()) {
		count++
	}
	l.backup()
	return count
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexQuery; l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions
const (
	colon      = ':'
	leftParen  = '('
	rightParen = ')'
	leftCurly  = '{'
	rightCurly = '}'
)

const (
	whitespace   = " \t\n\r"
	digits       = "0123456789"
)

func lexQuery(l *lexer) stateFn {
	if r := l.peek(); isSpace(r) {
		l.accept(whitespace)
		l.emit(itemSpace)
		return lexQuery

	} else if strings.HasPrefix(l.input[l.pos:], keywords[itemQueryKeyword]) {
		return lexQueryKeyword

	} else if r == eof {
		l.emit(itemEOF)
		return nil

	} else {
		return l.errorf("queries must begin with the query keyword")
	}
}

func lexQueryKeyword(l *lexer) stateFn {
	l.pos += Pos(len(keywords[itemQueryKeyword]))
	l.emit(itemQueryKeyword)

	// query must be followed by at least one whitespace
	if !l.accept(whitespace) {
		return l.errorf("query keyword must be followed by a whitespace")
	}

	l.acceptRun(whitespace) // skip any additional whitespace
	l.emit(itemSpace)

	return lexQueryName
}

func lexQueryName(l *lexer) stateFn {
	if length := l.acceptFn(isAlpha); length == 0 {
		return l.errorf("query name must begin with either an alphabet or an underscore")
	}
	l.acceptFn(isAlphaNumeric)
	l.emit(itemQueryName)

	// skip any whitespace
	if count := l.acceptRun(whitespace); count > 0 {
		l.emit(itemSpace)
	}

	return lexQueryBegin
}

func lexQueryBegin(l *lexer) stateFn {
	if l.next() != leftCurly {
		return l.errorf("queries must begin with a %s", leftCurly)
	}
	l.emit(itemQueryBegin)

	return lexInsideQuery
}

func lexInsideQuery(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexInsideQuery

	case isAlpha(r):
		return lexQueryField

	case r == rightCurly:
		l.next()
		l.emit(itemQueryEnd)
		return lexQuery

	default:
		return l.errorf("no field specified within query")
	}
}

// we've already peeked so we know the first character is an alpha
func lexQueryField(l *lexer) stateFn {
	l.acceptFn(isAlpha)
	l.acceptFn(isAlphaNumeric)
	l.emit(itemQueryField)

	return lexQueryArgs
}

func lexQueryArgs(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexQueryArgs

	case r == leftParen:
		return lexQueryLeftParen

	case r == leftCurly:
		return lexSelectorBegin

	default:
		return l.errorf("field name should be followed by arguments, whitespace, or a selector block")
	}
}

func lexQueryLeftParen(l *lexer) stateFn {
	if l.next() != leftParen {
		return l.errorf("query arguments must begin with a %s", leftParen)
	}
	l.emit(itemParamBegin)

	return lexInsideParam
}

func lexInsideParam(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexInsideParam

	case isAlpha(r):
		l.acceptFn(isAlpha)
		l.acceptFn(isAlphaNumeric)
		l.emit(itemParamName)
		return lexParamColon

	case r == rightParen:
		l.next()
		l.emit(itemParamEnd)
		return lexSelectorBegin

	default:
		return l.errorf("expected query argument")
	}
}

func lexParamColon(l *lexer) stateFn {
	if r := l.peek(); isSpace(r) {
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexParamColon

	} else if r == colon {
		l.next()
		l.emit(itemColon)
		return lexParamValue

	} else {
		return l.errorf("expected parameter name to be followed by a color")
	}
}

func lexParamValue(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexParamValue

	case isNumeric(r) || isSign(r):
		if !l.scanNumber() {
			return l.errorf("invalid number format for parameter")
		}
		l.emit(itemNumber)
		return lexInsideParam

	default:
		return l.errorf("unexpect value for parameter")
	}
}

func lexSelectorBegin(l *lexer) stateFn {
	if r := l.peek(); isSpace(r) {
		l.acceptRun(whitespace)
		l.emit(itemSpace)
	}

	if l.next() != leftCurly {
		return l.errorf("expected field selector to begin with %s", leftCurly)
	}
	l.emit(itemSelectorBegin)

	return lexInsideSelector
}

func lexInsideSelector(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexInsideSelector

	case isAlpha(r):
		l.acceptFn(isAlpha)
		l.acceptFn(isAlphaNumeric)
		l.emit(itemSelector)
		return lexInsideSelector

	case r == rightCurly:
		l.next()
		l.emit(itemSelectorEnd)
		return lexInsideQuery

	default:
		return l.errorf("unexpected value inside selector")
	}
}

func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")

	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}

	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}
	return true
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isAlpha reports whether r is an alphabetic or underscore
func isAlpha(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

// isNumeric reports whether r is numeric
func isNumeric(r rune) bool {
	return strings.IndexRune(digits, r) >= 0
}

// isNumeric reports whether r is a signed value
func isSign(r rune) bool {
	return r == '-' || r == '+'
}
