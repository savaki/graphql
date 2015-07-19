package mapops

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStore(t *testing.T) {
	Convey("Given a store", t, func() {
		friends := []string{
			"james",
			"jen",
			"jill",
			"joe",
		}
		data := map[string]interface{}{
			"bill": map[string]interface{}{
				"friends": friends,
			},
		}
		store := New(data)

		field, err := store.Query("bill")
		So(err, ShouldBeNil)
		So(field, ShouldNotBeNil)

		selection, err := field.Selection()
		So(err, ShouldBeNil)
		So(selection, ShouldNotBeNil)

		v, err := selection.Fetch("friends")
		So(err, ShouldBeNil)
		So(v, ShouldNotBeNil)

		value, err := v.Value()
		So(err, ShouldBeNil)
		So(value, ShouldNotBeNil)
		So(value, ShouldResemble, friends)
	})
}
