package parse

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLexSimple(t *testing.T) {
	Convey("Verify #lex on simple grammar", t, func() {
		l := lex("simple", `query sample {
			user(id: 4) {
				firstName
			}
		}`)

		wants := []item{
			{typ: itemQueryKeyword},
			{typ: itemSpace},
			{typ: itemQueryName, val: "sample"},
			{typ: itemSpace},
			{typ: itemQueryBegin},
			{typ: itemSpace},
			{typ: itemQueryField, val: "user"},
			{typ: itemParamBegin},
			{typ: itemParamName, val: "id"},
			{typ: itemColon},
			{typ: itemSpace},
			{typ: itemNumber, val: "4"},
			{typ: itemParamEnd},
			{typ: itemSpace},
			{typ: itemSelectorBegin},
			{typ: itemSpace},
			{typ: itemSelector, val: "firstName"},
			{typ: itemSpace},
			{typ: itemSelectorEnd},
			{typ: itemSpace},
			{typ: itemQueryEnd},
			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

func TestLexEmpty(t *testing.T) {
	Convey("Verify #lex on empty grammar", t, func() {
		l := lex("simple", ``)

		wants := []item{
			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

// @see https://news.ycombinator.com/item?id=8978936
/*
func TestLexHackerNews(t *testing.T) {
	Convey("Verify #lex on hn grammar", t, func() {
		l := lex("simple", `viewer() {
    posts {
      node {
        author { id, name, favorite_color },
        // any other post data you want
      }
    },
    friends {
      node {
        id,
        name,
        favorite_color,
      }
    },
    notifications {
      node {
        source { id, name, favorite_color },
        // any other notification fields you want
      }
    },
  }`)

		wants := []item{
			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}
*/

func VerifyWants(l *lexer, wants []item) {
	for _, want := range wants {
		item := l.nextItem()
		So(item.typ, ShouldEqual, want.typ)
		if want.val != "" {
			So(item.val, ShouldEqual, want.val)
		}
	}
}
