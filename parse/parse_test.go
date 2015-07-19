package parse

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseSimple(t *testing.T) {
	Convey("Given a simple graphql statement", t, func() {
		q := `query user(id:123) { close_friends: friends.max(5).distance(1) { picture } }`
		doc, err := Parse(q)
		So(err, ShouldBeNil)
		So(doc, ShouldNotBeNil)
	})
}
