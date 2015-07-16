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
	itemColonEquals                  // colon-equals (':=') introducing a declaration
	itemEOF
	itemField      // alphanumeric identifier starting with '.'
	itemIdentifier // alphanumeric identifier not starting with '.'
	itemLeftCurly  // left action delimiter
	itemLeftParen  // '(' inside action
	itemNumber     // simple number, including imaginary
	itemRawString  // raw quoted string (includes quotes)
	itemRightDelim // right action delimiter
	itemRightParen // ')' inside action
	itemSpace      // run of spaces separating arguments
	itemString     // quoted string (includes quotes)
	itemText       // plain text
	itemQueryName  // the name for a query
	itemQueryField // the field within the query

	// ONLY KEYWORDS BELOW THIS POINT
	itemKeyword      // used only to delimit the keywords
	itemDot          // the cursor, spelled '.'
	itemNil          // the untyped nil constant, easiest to treat as a keyword
	itemQueryKeyword // query keyword
)

var key = map[string]itemType{
	".":     itemDot,
	"nil":   itemNil,
	"query": itemQueryKeyword,
}

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
	leftCurly    = "{"
	rightCurly   = "}"
	leftComment  = "/*"
	rightComment = "*/"
	whitespace   = " \t\n"
	nameFirst    = "_ABCDEFGHIJKLMNOPQRSTV"
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

	return lexQueryLeftCurly
}

func lexQueryLeftCurly(l *lexer) stateFn {
	if !l.accept(leftCurly) {
		l.errorf("queries must begin with a %s", leftCurly)
	}
	l.emit(itemLeftCurly)

	return lexInsideQuery
}

func lexInsideQuery(l *lexer) stateFn {
	if r := l.peek(); isSpace(r) {
		l.accept(whitespace)
		l.emit(itemSpace)
		return lexInsideQuery

	} else if isAlpha(r) {
		return lexQueryField

	} else {
		return l.errorf("no field specified within query")
	}
}

// we've already peeked so we know the first character is an alpha
func lexQueryField(l *lexer) stateFn {
	l.acceptFn(isAlpha)
	l.acceptFn(isAlphaNumeric)
	l.emit(itemQueryField)

	// we're up to here
	l.emit(itemEOF)
	return nil
}

// lexComment scans a comment. The left comment marker is known to be present.
func lexComment(l *lexer) stateFn {
	l.pos += Pos(len(leftComment))
	i := strings.Index(l.input[l.pos:], rightComment)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i + len(rightComment))
	if !strings.HasPrefix(l.input[l.pos:], rightCurly) {
		return l.errorf("comment ends before closing delimiter")
	}
	l.pos += Pos(len(rightCurly))
	l.ignore()
	return lexQuery
}

// lexRightDelim scans the right delimiter, which is known to be present.
func lexRightDelim(l *lexer) stateFn {
	l.pos += Pos(len(rightCurly))
	l.emit(itemRightDelim)
	return lexQuery
}

// lexInsideAction scans the elements inside action delimiters.
func lexInsideAction(l *lexer) stateFn {
	// Either number, quoted string, or identifier.
	// Spaces separate arguments; runs of spaces turn into itemSpace.
	// Pipe symbols separate and are emitted.
	if strings.HasPrefix(l.input[l.pos:], rightCurly) {
		if l.parenDepth == 0 {
			return lexRightDelim
		}
		return l.errorf("unclosed left paren")
	}
	switch r := l.next(); {
	case r == eof || isEndOfLine(r):
		return l.errorf("unclosed action")
	case isSpace(r):
		return lexSpace
	case r == ':':
		if l.next() != '=' {
			return l.errorf("expected :=")
		}
		l.emit(itemColonEquals)
	case r == '"':
		return lexQuote
	case r == '`':
		return lexRawQuote
	case r == '\'':
		return lexChar
	case r == '.':
		// special look-ahead for ".field" so we don't break l.backup().
		if l.pos < Pos(len(l.input)) {
			r := l.input[l.pos]
			if r < '0' || '9' < r {
				return lexField
			}
		}
		fallthrough // '.' can start a number.
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case r == '(':
		l.emit(itemLeftParen)
		l.parenDepth++
		return lexInsideAction
	case r == ')':
		l.emit(itemRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren %#U", r)
		}
		return lexInsideAction
	case r <= unicode.MaxASCII && unicode.IsPrint(r):
		l.emit(itemChar)
		return lexInsideAction
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
	return lexInsideAction
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(itemSpace)
	return lexInsideAction
}

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
		// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch {
			case key[word] > itemKeyword:
				l.emit(key[word])
			case word[0] == '.':
				l.emit(itemField)
			case word == "true", word == "false":
				l.emit(itemBool)
			default:
				l.emit(itemIdentifier)
			}
			break Loop
		}
	}
	return lexInsideAction
}

// lexField scans a field: .Alphanumeric.
// The . has been scanned.
func lexField(l *lexer) stateFn {
	return lexFieldOrVariable(l, itemField)
}

// lexVariable scans a field or variable: [.$]Alphanumeric.
// The . or $ has been scanned.
func lexFieldOrVariable(l *lexer, typ itemType) stateFn {
	if l.atTerminator() { // Nothing interesting follows -> "." or "$".
		l.emit(itemDot)
		return lexInsideAction
	}
	var r rune
	for {
		r = l.next()
		if !isAlphaNumeric(r) {
			l.backup()
			break
		}
	}
	if !l.atTerminator() {
		return l.errorf("bad character %#U", r)
	}
	l.emit(typ)
	return lexInsideAction
}

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier. Breaks .X.Y into two pieces. Also catches cases
// like "$x+2" not being acceptable without a space, in case we decide one
// day to implement arithmetic.
func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	switch r {
	case eof, '.', ',', '|', ':', ')', '(':
		return true
	}
	// Does r start the delimiter? This can be ambiguous (with delim=="//", $x/2 will
	// succeed but should fail) but only in extremely rare cases caused by willfully
	// bad choice of delimiter.
	if rd, _ := utf8.DecodeRuneInString(rightCurly); rd == r {
		return true
	}
	return false
}

// lexChar scans a character constant. The initial quote is already
// scanned. Syntax checking is done by the parser.
func lexChar(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated character constant")
		case '\'':
			break Loop
		}
	}
	l.emit(itemCharConstant)
	return lexInsideAction
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	if sign := l.peek(); sign == '+' || sign == '-' {
		// Complex: 1+2i. No spaces, must end in 'i'.
		if !l.scanNumber() || l.input[l.pos-1] != 'i' {
			return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
		}
		l.emit(itemComplex)
	} else {
		l.emit(itemNumber)
	}
	return lexInsideAction
}
func (l *lexer) scanNumber() bool {
	// Optional leading sign.
	l.accept("+-")

	// Is it hex?
	digits := "0123456789"
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

// lexQuote scans a quoted string.
func lexQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}
	l.emit(itemString)
	return lexInsideAction
}

// lexRawQuote scans a raw quoted string.
func lexRawQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case eof, '\n':
			return l.errorf("unterminated raw quoted string")
		case '`':
			break Loop
		}
	}
	l.emit(itemRawString)
	return lexInsideAction
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isNameStart reports whether r is an alphabetic or underscore
func isAlpha(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}
