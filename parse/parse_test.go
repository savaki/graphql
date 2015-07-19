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

func TestParseSimple2(t *testing.T) {
	Convey("Verify #parse on empty grammar", t, func() {
		q := `query bill { friends }`
		doc, err := Parse(q)
		So(err, ShouldBeNil)
		So(doc, ShouldNotBeNil)
	})
}

func TestParseString(t *testing.T) {
	Convey("Verify #parse on nested grammar with string", t, func() {
		q := `query city: GET("http://api.openweathermap.org/data/2.5/weather?lat=35&lon=139") {
			name
			weather: main {
				temp: temperature
			}
		}`
		doc, err := Parse(q)
		So(err, ShouldBeNil)
		So(doc, ShouldNotBeNil)
	})
}
