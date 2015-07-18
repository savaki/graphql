package parse

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		l := lex("simple", `query sample {
			user(id: 4) {
				firstName
			}
		}`)
		for item := l.nextItem(); item.typ != itemEOF; item = l.nextItem() {
		}
	}
}

func TestLexComplex1(t *testing.T) {
	Convey("Verify #lex on complex grammar", t, func() {
		l := lex("complex",
			`query user(id:123) {
				close_friends: friends.max(5).distance(1) {
					picture
				}
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "user"},
			{typ: itemLeftParen},
			{typ: itemName, val: "id"},
			{typ: itemColon},
			{typ: itemNumber, val: "123"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemName, val: "close_friends"},
			{typ: itemColon},
			{typ: itemName, val: "friends"},
			{typ: itemDot},
			{typ: itemName, val: "max"},
			{typ: itemLeftParen},
			{typ: itemNumber, val: "5"},
			{typ: itemRightParen},
			{typ: itemDot},
			{typ: itemName, val: "distance"},
			{typ: itemLeftParen},
			{typ: itemNumber, val: "1"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemName, val: "picture"},
			{typ: itemRightCurly},
			{typ: itemRightCurly},
			{typ: itemEOF},
		}
		VerifyWants(l, wants)
	})
}

func TestLexSimple(t *testing.T) {
	Convey("Verify #lex on simple grammar", t, func() {
		l := lex("simple", `query sample {
			user(id: 4) {
				firstName
			}
		}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "sample"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "user"},
			{typ: itemLeftParen},
			{typ: itemName, val: "id"},
			{typ: itemColon},
			{typ: itemNumber, val: "4"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemName, val: "firstName"},
			{typ: itemRightCurly},
			{typ: itemRightCurly},
			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

func TestLexAlias(t *testing.T) {
	Convey("Verify #lex on simple grammar", t, func() {
		l := lex("simple", `query sample {
			me:user(id: 4) {
				first : firstName
			}
		}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "sample"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "me"},
			{typ: itemColon},
			{typ: itemName, val: "user"},
			{typ: itemLeftParen},
			{typ: itemName, val: "id"},
			{typ: itemColon},
			{typ: itemNumber, val: "4"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemName, val: "first"},
			{typ: itemColon},
			{typ: itemName, val: "firstName"},
			{typ: itemRightCurly},
			{typ: itemRightCurly},
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
