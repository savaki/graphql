// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package parse

import (
	"fmt"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	end Pos      // the length of this item
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
	itemError itemType = iota // error occurred; value is text of error
	itemBool                  // boolean constant
	itemEOF
	itemName       // alphanumeric identifier starting with '.'
	itemQuery      // query keyword
	itemLeftCurly  // marks start of query block
	itemRightCurly // right action delimiter
	itemLeftParen  // '(' inside action
	itemRightParen // ')' inside action
	itemNumber     // simple number, including imaginary
	itemSpace      // run of spaces separating arguments
	itemColon      // the : separating param name from param value
	itemComma      // the comma separating elements

	// ONLY KEYWORDS BELOW THIS POINT
	itemKeyword // used only to delimit the keywords
	itemDot     // the cursor, spelled '.'
	itemNil     // the untyped nil constant, easiest to treat as a keyword

	// TO BE REMOVED
	itemText         // plain text
	itemString       // quoted string (includes quotes)
	itemRawString    // raw quoted string (includes quotes)
	itemIdentifier   // alphanumeric identifier not starting with '.'
	itemChar         // printable ASCII character; grab bag for comma etc.
	itemCharConstant // character constant
	itemComplex      // complex constant (1+2i); imaginary is just a number
)

var keywords = map[itemType]string{
	itemQuery: "query",
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name    string      // the name of the input; used only for error reports
	input   string      // the string being scanned
	state   stateFn     // the next lexing function to enter
	pos     Pos         // current position in the input
	start   Pos         // start position of this item
	width   Pos         // width of last rune read from input
	lastPos Pos         // position of most recent item returned by nextItem
	items   chan item   // channel of scanned items
	depth   int         // selector depth
	token   [2]itemType // three-token look behind for parser.
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
	l.items <- item{t, l.start, l.pos, l.input[l.start:l.pos]}
	l.start = l.pos

	if t != itemSpace {
		// three token look behind
		l.token[1] = l.token[0]
		l.token[0] = t
	}
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

func (l *lexer) acceptOrdered(valid string) bool {
	for index, r := range valid {
		if l.next() != r {
			for i := 0; i <= index; i++ {
				l.backup()
			}
			return false
		}
	}
	return true
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
	log.Printf(format, args...)
	l.items <- item{itemError, l.start, 0, fmt.Sprintf(format, args...)}
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
	for l.state = lexRoot; l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions
const (
	dot        = '.'
	colon      = ':'
	plus       = '+'
	minus      = '-'
	comma      = ','
	leftParen  = '('
	rightParen = ')'
	leftCurly  = '{'
	rightCurly = '}'
)

const (
	whitespace = " \t\n\r"
	digits     = "0123456789"
)

func lexRoot(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.accept(whitespace)
		l.emit(itemSpace)
		return lexRoot

	case strings.HasPrefix(l.input[l.pos:], keywords[itemQuery]):
		return lexQuery

	case r == eof:
		l.emit(itemEOF)
		return nil

	default:
		return l.errorf("queries must begin with the query keyword ->")
	}
}

func lexQuery(l *lexer) stateFn {
	l.pos += Pos(len(keywords[itemQuery]))
	l.emit(itemQuery)

	// query must be followed by at least one whitespace
	if !isSpace(l.peek()) {
		return l.errorf("query keyword must be followed by a whitespace")
	}

	return lexField
}

func lexField(l *lexer) stateFn {
	if l.acceptRun(whitespace) > 0 {
		l.emit(itemSpace)
	}

	// scan field
	if !l.scanField() {
		return l.errorf("fields must begin with either an alpha or an underscore")
	}

	return lexFieldFilter
}

func lexFieldFilter(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexFieldFilter

	case r == dot && l.token[0] == itemName:
		l.next()
		l.emit(itemDot)
		return lexField

	case r == leftParen && l.token[0] == itemName:
		l.next()
		l.emit(itemLeftParen)
		return lexFieldArg

	case r == colon && l.token[0] == itemName && l.token[1] != itemColon:
		l.next()
		l.emit(itemColon)
		return lexField

	case r == leftCurly:
		l.next()
		l.emit(itemLeftCurly)
		l.depth++
		return lexField

	case r == rightCurly && l.depth > 0:
		l.next()
		l.emit(itemRightCurly)
		l.depth--
		return lexPostField

	case r == eof:
		l.emit(itemEOF)
		return nil

	default:
		log.Printf("-> %s\n", l.input[l.start:])
		return l.errorf("expected selector")
	}
}

func lexPostField(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexFieldFilter

	default:
		l.emit(itemEOF)
		return nil
	}
}

func lexFieldArg(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexFieldArg

	case r == plus || r == minus || isNumeric(r):
		if !l.scanNumber() {
			return l.errorf("invalid number format for arg")
		}
		return lexFieldNext

	case r == 't' || r == 'f':
		if !l.scanBool() {
			return l.errorf("invalid bool format for arg")
		}
		return lexFieldNext

	case r == rightParen:
		l.next()
		l.emit(itemRightParen)
		return lexFieldFilter

	case r == comma && l.token[0] == itemName:
		l.next()
		l.emit(itemComma)
		return lexFieldArg

	case r == colon && l.token[0] == itemName:
		l.next()
		l.emit(itemColon)
		return lexFieldValue

	case isAlpha(r):
		l.next()
		l.acceptFn(isAlphaNumeric)
		l.emit(itemName)
		return lexFieldNext

	default:
		return l.errorf("unexpected field arg")
	}
}

func lexFieldValue(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexFieldValue

	case r == plus || r == minus || isNumeric(r):
		if !l.scanNumber() {
			return l.errorf("invalid number format for arg")
		}
		return lexFieldNext

	case r == 't' || r == 'f':
		if !l.scanBool() {
			return l.errorf("invalid bool format for arg")
		}
		return lexFieldNext

	default:
		return l.errorf("unexpected field value")
	}
}

// lexFieldNext scans the content immediately after a field name
func lexFieldNext(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isSpace(r):
		l.acceptRun(whitespace)
		l.emit(itemSpace)
		return lexFieldNext

	case r == comma:
		l.next()
		l.emit(itemComma)
		return lexFieldArg

	case r == colon && l.token[0] == itemName:
		l.next()
		l.emit(itemColon)
		return lexFieldValue

	case r == rightParen:
		l.next()
		l.emit(itemRightParen)
		return lexFieldFilter

	default:
		return l.errorf("argument or ) expected")
	}
}

func (l *lexer) scanField() bool {
	if length := l.acceptFn(isAlpha); length == 0 {
		return false
	}
	l.acceptFn(isAlphaNumeric)
	l.emit(itemName)
	return true
}

func (l *lexer) scanBool() bool {
	switch {
	case l.acceptOrdered("true") || l.acceptOrdered("false"):
		l.emit(itemBool)
		return true

	default:
		return false
	}
}

func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")

	length := l.acceptRun(digits)
	if l.accept(".") {
		length = length + l.acceptRun(digits)
	}

	if length == 0 {
		// no digits
		return false
	}

	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}

	l.emit(itemNumber)
	return true
}

// isSpace reports whether r is a space character.
// space characters are: space, tab, carriage-return (\r), line feed (\n), and comma
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ','
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
