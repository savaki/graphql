package parse

import (
	"encoding/json"
	"testing"

	"fmt"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseSimple(t *testing.T) {
	Convey("Given a simple graphql statement", t, func() {
		q := `query user(id:123) { close_friends: friends.max(5).distance(1) { picture } }`
		node, err := Parse(q)
		So(err, ShouldBeNil)
		So(node, ShouldNotBeNil)

		data, _ := json.MarshalIndent(node, "", "..")
		fmt.Println(string(data))
	})
}
