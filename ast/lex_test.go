package ast

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

func TestLexHello(t *testing.T) {
	Convey("Verify #lex on hello world", t, func() {
		l := lex("hello world", `{hello}`)
		wants := []item{
			{typ: itemLeftCurly},
			{typ: itemName, val: "hello"},
			{typ: itemRightCurly},
			{typ: itemEOF},
		}
		VerifyWants(l, wants)
	})
}

func TestLexComplex1(t *testing.T) {
	Convey("Verify #lex on complex grammar", t, func() {
		l := lex("complex",
			`query user(id:123) {
				close_friends: friends(max: 5, distance: 1.2) {
					picture
				}
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "user"},
			{typ: itemLeftParen},
			{typ: itemName, val: "id"},
			{typ: itemColon},
			{typ: itemIntValue, val: "123"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemName, val: "close_friends"},
			{typ: itemColon},
			{typ: itemName, val: "friends"},
			{typ: itemLeftParen},
			{typ: itemName, val: "max"},
			{typ: itemColon},
			{typ: itemIntValue, val: "5"},
			{typ: itemName, val: "distance"},
			{typ: itemColon},
			{typ: itemFloatValue, val: "1.2"},
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
		l := lex("simple", `
			query sample {
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
			{typ: itemIntValue, val: "4"},
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

func TestLexDefaultValue(t *testing.T) {
	Convey("Verify #lex on grammar with default value", t, func() {
		l := lex("default value", `
			query sample($id: Int = 5) {
				user {
					firstName
				}
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "sample"},
			{typ: itemLeftParen},
			{typ: itemVariable, val: "id"},
			{typ: itemColon},
			{typ: itemIntType},
			{typ: itemEqual},
			{typ: itemIntValue, val: "5"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemName, val: "user"},
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
		l := lex("simple", `
			query sample {
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
			{typ: itemIntValue, val: "4"},
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

func TestLexVariables(t *testing.T) {
	Convey("Verify #lex on grammar with variables", t, func() {
		l := lex("variables", `
			query getZuckProfile($devicePicSize: Int) {
				user(id: 4) {
					id
					name
					profilePic(size: $devicePicSize)
				}
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "getZuckProfile"},
			{typ: itemLeftParen},
			{typ: itemVariable, val: "devicePicSize"},
			{typ: itemColon},
			{typ: itemIntType},
			{typ: itemRightParen},
			{typ: itemLeftCurly},

			{typ: itemName, val: "user"},
			{typ: itemLeftParen},
			{typ: itemName, val: "id"},
			{typ: itemColon},
			{typ: itemIntValue, val: "4"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},

			{typ: itemName, val: "id"},
			{typ: itemName, val: "name"},
			{typ: itemName, val: "profilePic"},
			{typ: itemLeftParen},
			{typ: itemName, val: "size"},
			{typ: itemColon},
			{typ: itemVariable, val: "devicePicSize"},
			{typ: itemRightParen},

			{typ: itemRightCurly},
			{typ: itemRightCurly},

			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

func TestLexDirective(t *testing.T) {
	Convey("Verify #lex on grammar with fragments and conditionals", t, func() {
		l := lex("variables", `
			query hasConditionalFragment($condition: Boolean) {
			  ...maybeFragment @include(if: $condition) @solo
			}
			fragment maybeFragment on Query @include(if: $condition) {
			  me {
				name
			  }
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "hasConditionalFragment"},
			{typ: itemLeftParen},
			{typ: itemVariable, val: "condition"},
			{typ: itemColon},
			{typ: itemBooleanType},
			{typ: itemRightParen},
			{typ: itemLeftCurly},

			{typ: itemEllipses},
			{typ: itemName, val: "maybeFragment"},
			{typ: itemAtSign},
			{typ: itemName, val: "include"},
			{typ: itemLeftParen},
			{typ: itemName, val: "if"},
			{typ: itemColon},
			{typ: itemVariable, val: "condition"},
			{typ: itemRightParen},
			{typ: itemAtSign},
			{typ: itemName, val: "solo"},
			{typ: itemRightCurly},

			{typ: itemFragment},
			{typ: itemName, val: "maybeFragment"},
			{typ: itemOn},
			{typ: itemName, val: "Query"},
			{typ: itemAtSign},
			{typ: itemName, val: "include"},
			{typ: itemLeftParen},
			{typ: itemName, val: "if"},
			{typ: itemColon},
			{typ: itemVariable, val: "condition"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},

			{typ: itemName, val: "me"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "name"},
			{typ: itemRightCurly},

			{typ: itemRightCurly},

			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

func TestLexSimple2(t *testing.T) {
	Convey("Verify #lex on empty grammar", t, func() {
		l := lex("simple", `query bill { friends }`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "bill"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "friends"},
			{typ: itemRightCurly},
		}

		VerifyWants(l, wants)
	})
}

func TestLexString(t *testing.T) {
	Convey("Verify #lex on empty grammar", t, func() {
		l := lex("simple", `
query city: GET(url: "http://api.openweathermap.org/data/2.5/weather?lat=35&lon=139") {
	name
	weather: main {
		temp: temperature
	}
}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "city"},
			{typ: itemColon},
			{typ: itemName, val: "GET"},
			{typ: itemLeftParen},
			{typ: itemName, val: "url"},
			{typ: itemColon},
			{typ: itemStringValue, val: "http://api.openweathermap.org/data/2.5/weather?lat=35&lon=139"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemName, val: "name"},
			{typ: itemName, val: "weather"},
			{typ: itemColon},
			{typ: itemName, val: "main"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "temp"},
			{typ: itemColon},
			{typ: itemName, val: "temperature"},
			{typ: itemRightCurly},
			{typ: itemRightCurly},
		}

		VerifyWants(l, wants)
	})
}

func TestLexFragment(t *testing.T) {
	Convey("Verify we can parse the sample fragmnt", t, func() {
		l := lex("fragment", `
			query withFragments {
			  user(id: 4) {
				friends(first: 10) {
				  ...friendFields
				}
				mutualFriends(first: 10) {
				  ...friendFields
				}
			  }
			}

			fragment friendFields on User {
			  id
			  name
			  profilePic(size: 50)
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "withFragments"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "user"},
			{typ: itemLeftParen},
			{typ: itemName, val: "id"},
			{typ: itemColon},
			{typ: itemIntValue, val: "4"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},

			{typ: itemName, val: "friends"},
			{typ: itemLeftParen},
			{typ: itemName, val: "first"},
			{typ: itemColon},
			{typ: itemIntValue, val: "10"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemEllipses},
			{typ: itemName, val: "friendFields"},
			{typ: itemRightCurly},

			{typ: itemName, val: "mutualFriends"},
			{typ: itemLeftParen},
			{typ: itemName, val: "first"},
			{typ: itemColon},
			{typ: itemIntValue, val: "10"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemEllipses},
			{typ: itemName, val: "friendFields"},
			{typ: itemRightCurly},

			{typ: itemRightCurly},
			{typ: itemRightCurly},

			{typ: itemFragment},
			{typ: itemName, val: "friendFields"},
			{typ: itemOn},
			{typ: itemName, val: "User"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "id"},
			{typ: itemName, val: "name"},
			{typ: itemName, val: "profilePic"},
			{typ: itemLeftParen},
			{typ: itemName, val: "size"},
			{typ: itemColon},
			{typ: itemIntValue, val: "50"},
			{typ: itemRightParen},
			{typ: itemRightCurly},
			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

func TestLexNestedFragments(t *testing.T) {
	Convey("Verify we can parse the sample fragmnt", t, func() {
		l := lex("fragment", `
			query withNestedFragments {
			  user(id: 4) {
				friends(first: 10) {
				  ...friendFields
				}
				mutualFriends(first: 10) {
				  ...friendFields
				}
			  }
			}

			fragment friendFields on User {
			  id
			  name
			  ...standardProfilePic
			}

			fragment standardProfilePic on User {
			  profilePic(size: 50)
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "withNestedFragments"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "user"},
			{typ: itemLeftParen},
			{typ: itemName, val: "id"},
			{typ: itemColon},
			{typ: itemIntValue, val: "4"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},

			{typ: itemName, val: "friends"},
			{typ: itemLeftParen},
			{typ: itemName, val: "first"},
			{typ: itemColon},
			{typ: itemIntValue, val: "10"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemEllipses},
			{typ: itemName, val: "friendFields"},
			{typ: itemRightCurly},

			{typ: itemName, val: "mutualFriends"},
			{typ: itemLeftParen},
			{typ: itemName, val: "first"},
			{typ: itemColon},
			{typ: itemIntValue, val: "10"},
			{typ: itemRightParen},
			{typ: itemLeftCurly},
			{typ: itemEllipses},
			{typ: itemName, val: "friendFields"},
			{typ: itemRightCurly},

			{typ: itemRightCurly},
			{typ: itemRightCurly},

			{typ: itemFragment},
			{typ: itemName, val: "friendFields"},
			{typ: itemOn},
			{typ: itemName, val: "User"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "id"},
			{typ: itemName, val: "name"},
			{typ: itemEllipses},
			{typ: itemName, val: "standardProfilePic"},
			{typ: itemRightCurly},

			{typ: itemFragment},
			{typ: itemName, val: "standardProfilePic"},
			{typ: itemOn},
			{typ: itemName, val: "User"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "profilePic"},
			{typ: itemLeftParen},
			{typ: itemName, val: "size"},
			{typ: itemColon},
			{typ: itemIntValue, val: "50"},
			{typ: itemRightParen},
			{typ: itemRightCurly},

			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

// @see https://news.ycombinator.com/item?id=8978936
func TestLexHackerNews(t *testing.T) {
	Convey("Verify #lex on hn grammar", t, func() {
		l := lex("simple", `
			query viewer() {
				posts {
					node {
						author { id, name, favorite_color },
						# any other post data you want
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
						# any other notification fields you want
					}
				},
			}`)

		wants := []item{
			{typ: itemQuery},
			{typ: itemName, val: "viewer"},
			{typ: itemLeftParen},
			{typ: itemRightParen},
			{typ: itemLeftCurly},

			// posts
			{typ: itemName, val: "posts"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "node"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "author"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "id"},
			{typ: itemName, val: "name"},
			{typ: itemName, val: "favorite_color"},
			{typ: itemRightCurly},
			{typ: itemRightCurly},
			{typ: itemRightCurly},

			// friends
			{typ: itemName, val: "friends"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "node"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "id"},
			{typ: itemName, val: "name"},
			{typ: itemName, val: "favorite_color"},
			{typ: itemRightCurly},
			{typ: itemRightCurly},

			// notifications
			{typ: itemName, val: "notifications"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "node"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "source"},
			{typ: itemLeftCurly},
			{typ: itemName, val: "id"},
			{typ: itemName, val: "name"},
			{typ: itemName, val: "favorite_color"},
			{typ: itemRightCurly},
			{typ: itemRightCurly},
			{typ: itemRightCurly},

			{typ: itemRightCurly},
			{typ: itemEOF},
		}

		VerifyWants(l, wants)
	})
}

func VerifyWants(l *lexer, wants []item) {
	for _, want := range wants {
		actual := l.nextItem()
		So(actual.typ, ShouldEqual, want.typ)
		if want.val != "" {
			So(actual.val, ShouldEqual, want.val)
		}
	}
}
