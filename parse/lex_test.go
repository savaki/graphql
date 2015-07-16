package parse

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Matcher struct {
	ItemType itemType
	Value    string
}

func (m Matcher) Match(item item) {
	So(item.itemType, ShouldEqual, m.ItemType)
	if m.Value != "" {
		So(item.val, ShouldEqual, m.Value)
	}
}

type Expectations struct {
	Matchers []Matcher
}

func expect(itemType itemType) *Expectations {
	return expectValue(itemType, "")
}

func expectValue(itemType itemType, value string) *Expectations {
	entry := Matcher{
		ItemType: itemType,
		Value:    value,
	}
	return &Expectations{
		Matchers: []Matcher{entry},
	}
}

func (e *Expectations) andThen(itemType itemType) *Expectations {
	return e.andThenValue(itemType, "")
}

func (e *Expectations) andThenValue(itemType itemType, value string) *Expectations {
	matcher := Matcher{
		ItemType: itemType,
		Value:    value,
	}
	return &Expectations{
		Matchers: append(e.Matchers, matcher),
	}
}

func (e *Expectations) Verify(l *lexer) {
	for _, matcher := range e.Matchers {
		item := l.nextItem()
		matcher.Match(item)
	}
}

func TestLexQuery(t *testing.T) {
	type expectation struct {
		ItemType itemType
		Value    string
	}

	type testCase struct {
		Text         string
		Expectations *Expectations
	}

	tests := []testCase{
		{"query sample { user(", expect(itemQueryKeyword).
			andThen(itemSpace).
			andThenValue(itemQueryName, "sample").
			andThen(itemSpace).
			andThen(itemLeftCurly).
			andThen(itemSpace).
			andThenValue(itemQueryField, "user").
			andThen(itemEOF)},

		{"querysample", expect(itemQueryKeyword).
			andThen(itemError)},

		{"", expect(itemEOF)},
	}

	Convey("Verify the lexer parses graphql", t, func() {
		for _, test := range tests {
			l := lex("sample", test.Text)
			test.Expectations.Verify(l)
		}
	})
}
