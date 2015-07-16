package parse

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLexQuery(t *testing.T) {
	type testCase struct {
		text  string
		wants []item
	}

	tests := []testCase{
		{"query sample { user(", []item{
			{typ: itemQueryKeyword},
			{typ: itemSpace},
			{typ: itemQueryName, val: "sample"},
			{typ: itemSpace},
			{typ: itemLeftCurly},
			{typ: itemSpace},
			{typ: itemQueryField, val: "user"},
			{typ: itemEOF},
		}},

		{"querysample", []item{
			{typ: itemQueryKeyword},
			{typ: itemError},
		}},

		{"", []item{
			{typ: itemEOF},
		}},
	}

	Convey("Verify the lexer parses graphql", t, func() {
		for _, test := range tests {
			l := lex("sample", test.text)
			VerifyWants(l, test.wants)
		}
	})
}

func VerifyWants(l *lexer, wants []item) {
	for _, want := range wants {
		item := l.nextItem()
		So(item.typ, ShouldEqual, want.typ)
		if want.val != "" {
			So(item.val, ShouldEqual, want.val)
		}
	}
}
