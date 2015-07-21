// Copyright 2015 Matt Ho. All rights reserved.
//
// Original source courtesy the Go Team 20111
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package grammar

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Pos int

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
	itemEOF
	itemName       // named item
	itemVariable   // variable
	itemLeftCurly  // marks start of query block
	itemRightCurly // right action delimiter
	itemLeftParen  // '(' inside action
	itemRightParen // ')' inside action
	itemAtSign     // '@' directive
	itemColon      // the : separating param name from param value
	itemComma      // the comma separating elements
	itemDot        // the cursor, spelled '.'
	itemNil        // the untyped nil constant, easiest to treat as a keyword
	itemEqual      // equal sign

	itemIntValue    // integer
	itemStringValue // string
	itemFloatValue  // floating point number

	// ONLY KEYWORDS BELOW THIS POINT
	itemKeyword     // used only to delimit the keywords
	itemQuery       // query keyword
	itemMutation    // mutations keyword
	itemFragment    // fragment keyword
	itemEllipses    // fragment definition, '...'
	itemTrue        // true
	itemFalse       // false
	itemOn          // fragment keyword
	itemIntType     // represents abstract Int type
	itemFloatType   // represents abstract Float type
	itemBooleanType // represents abstract Boolean type
	itemEnumType    // represents abstract Enum type
	itemArrayType   // represents abstract Array type
	itemObjectType  // represents abstract Object type
)

var keywords = map[itemType]string{
	itemQuery:       "query",
	itemMutation:    "mutation",
	itemFragment:    "fragment",
	itemEllipses:    "...",
	itemTrue:        "true",
	itemFalse:       "false",
	itemOn:          "on",
	itemIntType:     "Int",
	itemFloatType:   "Float",
	itemBooleanType: "Boolean",
	itemEnumType:    "Enum",
	itemArrayType:   "Array",
	itemObjectType:  "Object",
}

var allTypes = []itemType{
	itemIntType,
	itemFloatType,
	itemBooleanType,
	itemEnumType,
	itemArrayType,
	itemObjectType,
}

var unicodeMatcher = regexp.MustCompile(`u([0-9A-Fa-f]){4}`)

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
	token   [2]itemType // two-token look behind for parser.
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

	// two token look behind
	l.token[1] = l.token[0]
	l.token[0] = t
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

// accept consumes a 5 digit unicode value u/[0-9A-Fa-f]{4}/
func (l *lexer) acceptUnicode() bool {
	if !unicodeMatcher.MatchString(l.input[l.pos:]) {
		l.backup()
		return false
	}

	// unicode characters are 5 characters beginning from the u
	l.next()
	l.next()
	l.next()
	l.next()
	l.next()

	return true
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
	for l.state = lexDocument; l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions
const (
	dollar      = '$'
	atSign      = '@'
	dot         = '.'
	colon       = ':'
	plus        = '+'
	minus       = '-'
	equalSign   = '='
	doubleQuote = '"'
	escape      = '\\'
	comma       = ','
	leftParen   = '('
	rightParen  = ')'
	leftCurly   = '{'
	rightCurly  = '}'
)

const (
	whitespace       = ", \t\n\r"
	lineTerminator   = "\n\r"
	digits           = "0123456789"
	escapeCharacters = `"\/bfnrt` // see - https://github.com/facebook/graphql/blob/master/Section%208%20--%20Grammar.md
)

func lexDocument(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexDocument)

	case isComment(r):
		return l.ignoreComment(lexDocument)

	case l.hasPrefix(keywords[itemFragment]):
		return lexFragment

	case l.hasPrefix(keywords[itemQuery]):
		return lexQuery

	case l.hasPrefix(keywords[itemMutation]):
		return lexMutation

	case r == leftCurly:
		return lexSelectionSet

	case r == eof:
		l.emit(itemEOF)
		return nil

	default:
		return l.errorf("queries must begin with the query keyword ->")
	}
}

// lexQuery assumes the buffer begins with the query keyword
func lexQuery(l *lexer) stateFn {
	l.acceptOrdered(keywords[itemQuery])
	l.emit(itemQuery)

	// must be followed by at least one whitespace or comment
	if r := l.peek(); !isWhitespace(r) && !isComment(r) {
		return l.errorf("query keyword must be followed by either a whitespace or comment")
	}

	return lexField
}

// lexMutation assumes the buffer begins with the mutation keyword
func lexMutation(l *lexer) stateFn {
	l.acceptOrdered(keywords[itemMutation])
	l.emit(itemMutation)

	// must be followed by at least one whitespace or comment
	if r := l.peek(); !isWhitespace(r) && !isComment(r) {
		return l.errorf("query keyword must be followed by either a whitespace or comment")
	}

	return lexField
}

func lexField(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexField)

	case isComment(r):
		return l.ignoreComment(lexField)

	case isAlpha(r):
		return l.scanField(lexAfterField)

	default:
		return l.errorf("expected character for operation name")
	}
}

func lexAfterField(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexAfterField)

	case isComment(r):
		return l.ignoreComment(lexAfterField)

	case r == atSign:
		return lexDirective

	case r == leftParen:
		l.next()
		l.emit(itemLeftParen)
		return lexArgument

	case r == leftCurly:
		return lexSelectionSet

	case r == rightCurly:
		return lexEndSelection

	case r == colon && l.token[0] == itemName && l.token[1] != itemColon:
		l.next()
		l.emit(itemColon)
		return lexField

	case l.hasPrefix(keywords[itemEllipses]):
		l.acceptOrdered(keywords[itemEllipses])
		l.emit(itemEllipses)
		return lexAfterField

	case isAlpha(r) && l.token[0] == itemEllipses:
		return l.scanField(lexAfterField)

	case isAlpha(r):
		return lexField

	default:
		return l.errorf("unexpected values after field")
	}
}

func lexDirective(l *lexer) stateFn {
	// @
	if r := l.peek(); r != atSign {
		return l.errorf("directives must begin with an '@'")
	}
	l.next()
	l.emit(itemAtSign)

	// name
	if r := l.peek(); !isAlpha(r) {
		return l.errorf("directive @ sign must be immediately followed by an alpha")
	}
	l.acceptFn(isAlpha)
	l.acceptFn(isAlphaNumeric)
	l.emit(itemName)

	// (
	if r := l.peek(); r != leftParen {
		return l.errorf("directives must be followed by arguments")
	}
	l.next()
	l.emit(itemLeftParen)

	return lexArgument
}

func lexArgument(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexArgument)

	case isComment(r):
		return l.ignoreComment(lexArgument)

	case isAlpha(r):
		return l.scanField(lexColon)

	case r == dollar:
		return l.scanVariable(lexColon)

	case r == rightParen:
		l.next()
		l.emit(itemRightParen)
		return lexAfterField

	default:
		return l.errorf("unexpected argument")
	}
}

func lexColon(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexColon)

	case isComment(r):
		return l.ignoreComment(lexColon)

	case r == colon:
		l.next()
		l.emit(itemColon)
		return l.scanValue(lexDefaultValue)

	default:
		return l.errorf("expected colon")
	}
}

func lexDefaultValue(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexDefaultValue)

	case isComment(r):
		return l.ignoreComment(lexDefaultValue)

	case r == equalSign:
		l.next()
		l.emit(itemEqual)
		return l.scanValue(lexArgument)

	default:
		return lexArgument
	}
}

func lexSelectionSet(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexSelectionSet)

	case isComment(r):
		return l.ignoreComment(lexSelectionSet)

	case r == leftCurly:
		l.depth++
		l.next()
		l.emit(itemLeftCurly)
		return lexAfterField

	default:
		return l.errorf("expected begin selection")
	}
}

func lexEndSelection(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexEndSelection)

	case isComment(r):
		return l.ignoreComment(lexEndSelection)

	case r == rightCurly:
		l.depth--
		l.next()
		l.emit(itemRightCurly)
		if l.depth == 0 {
			return lexDocument
		} else {
			return lexAfterField
		}

	default:
		return l.errorf("expected right curly")
	}
}

func lexFragment(l *lexer) stateFn {
	l.pos += Pos(len(keywords[itemFragment]))
	l.emit(itemFragment)

	// fragment must be followed by at least one whitespace or comment
	if r := l.peek(); !isWhitespace(r) && !isComment(r) {
		return l.errorf("fragment keyword must be followed by a whitespace or a comment")
	}

	return lexFragmentName
}

func lexFragmentName(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexFragmentName)

	case isComment(r):
		return l.ignoreComment(lexFragmentName)

	case isAlpha(r):
		return l.scanField(lexOn)

	default:
		return l.errorf("expected right curly")
	}
}

func lexOn(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexOn)

	case isComment(r):
		return l.ignoreComment(lexOn)

	case strings.HasPrefix(l.input[l.pos:], keywords[itemOn]):
		l.acceptOrdered(keywords[itemOn])
		l.emit(itemOn)
		return lexFragmentType

	default:
		return l.errorf("expected on keyword")
	}
}

func lexFragmentType(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexFragmentType)

	case isComment(r):
		return l.ignoreComment(lexFragmentType)

	case isAlpha(r):
		return l.scanField(lexAfterFragmentType)

	default:
		return l.errorf("expected fragment type to be alpha numeric")
	}
}

func lexAfterFragmentType(l *lexer) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		return l.ignoreWhitespace(lexAfterFragmentType)

	case isComment(r):
		return l.ignoreComment(lexAfterFragmentType)

	case r == atSign:
		return lexDirective

	case r == leftCurly:
		return lexSelectionSet

	default:
		return l.errorf("expected fragment type to be alpha numeric")
	}
}

func (l *lexer) hasPrefix(word string) bool {
	return strings.HasPrefix(l.input[l.pos:], word)
}

func (l *lexer) ignoreWhitespace(fn stateFn) stateFn {
	l.acceptRun(whitespace)
	l.ignore()
	return fn
}

func (l *lexer) ignoreComment(fn stateFn) stateFn {
	if r := l.peek(); !isComment(r) {
		l.errorf("ignoreComment expects to start with the comment character")
	}

	l.next()
	l.acceptFn(isNotLineTerminator)
	l.ignore()
	return fn
}

func (l *lexer) scanValue(fn stateFn) stateFn {
	r := l.peek()
	switch {
	case isWhitespace(r):
		l.acceptRun(whitespace)
		l.ignore()
		return l.scanValue(fn)

	case isComment(r):
		l.next()
		l.acceptFn(isNotLineTerminator)
		l.ignore()
		return l.scanValue(fn)

	case r == doubleQuote:
		return l.scanString(fn)

	case r == plus || r == minus || isNumeric(r):
		return l.scanNumber(fn)

	case r == dollar:
		return l.scanVariable(fn)

	case strings.HasPrefix(l.input[l.pos:], keywords[itemTrue]):
		l.acceptOrdered("true")
		l.emit(itemTrue)
		return fn

	case strings.HasPrefix(l.input[l.pos:], keywords[itemFalse]):
		l.acceptOrdered("false")
		l.emit(itemFalse)
		return fn

	case isAlpha(r):
		return l.scanType(fn)

	default:
		return l.errorf("illegal value")
	}
}

func (l *lexer) scanField(fn stateFn) stateFn {
	if r := l.peek(); !isAlpha(r) {
		return l.errorf("invalid field; fields must start with an alpha character")
	}

	l.acceptFn(isAlpha)
	l.acceptFn(isAlphaNumeric)
	l.emit(itemName)
	return fn
}

func (l *lexer) scanVariable(fn stateFn) stateFn {
	if r := l.peek(); r != dollar {
		return l.errorf("invalid variable; variabls must start with a $")
	}
	l.next()
	l.ignore()

	if l.acceptFn(isAlpha) == 0 {
		return l.errorf("invalid variable; $ must be followed by an alpha")
	}
	l.acceptFn(isAlphaNumeric)
	l.emit(itemVariable)

	return fn
}

func (l *lexer) scanString(fn stateFn) stateFn {
	if r := l.peek(); r != doubleQuote {
		return l.errorf("strings must begin with a %v", doubleQuote)
	}

	l.next()
	l.ignore()

	for {
		switch l.peek() {
		case escape:
			l.next()
			if r := l.peek(); isEscapedCharacter(r) {
				l.next()

			} else if l.acceptUnicode() {
				continue

			} else {
				return l.errorf("invalid escape sequence")
			}

		case doubleQuote:
			l.emit(itemStringValue)
			l.next()
			l.ignore()
			return fn

		case eof:
			return l.errorf("unmatched double quotes")

		default:
			l.next()
		}
	}
}

func (l *lexer) scanType(fn stateFn) stateFn {
	for _, typ := range allTypes {
		if l.acceptOrdered(keywords[typ]) {
			l.emit(typ)
			return fn
		}
	}

	return l.errorf("unexpected type")
}

func (l *lexer) scanNumber(fn stateFn) stateFn {
	// Optional leading sign.
	l.accept("+-")

	typ := itemIntValue

	length := l.acceptRun(digits)
	if l.accept(".") {
		typ = itemFloatValue
		length = length + l.acceptRun(digits)
	}

	if length == 0 {
		// no digits
		return l.errorf("digits must be at least 0")
	}

	// Next thing mustn't be alphanumeric.
	if r := l.peek(); isAlpha(r) {
		return l.errorf("numbers may not immediately be followed by alphas")
	}

	l.emit(typ)
	return fn
}

func isComment(r rune) bool {
	return r == '#'
}

// isWhitespace reports whether r is a space character.
// space characters are: space, tab, carriage-return (\r), line feed (\n), and comma
func isWhitespace(r rune) bool {
	return strings.IndexRune(whitespace, r) >= 0
}

// isEndOfLine reports whether r is an end-of-line character.
func isLineTerminator(r rune) bool {
	return strings.IndexRune(lineTerminator, r) >= 0
}

func isNotLineTerminator(r rune) bool {
	return !isLineTerminator(r)
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

// isEscapeCharacter reports whether r is a valid escape character to follow a \
func isEscapedCharacter(r rune) bool {
	return strings.IndexRune(escapeCharacters, r) >= 0
}
