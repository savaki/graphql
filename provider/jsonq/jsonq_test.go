package jsonq

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Given a json block", t, func() {
		data := []byte(`{
			"a": {
				"name": "Adam",
				"age": 123
			}
	 	}`)

		store, err := New(data)
		So(err, ShouldBeNil)

		field, err := store.Fetch("a")
		So(err, ShouldBeNil)

		selection, err := field.Selection()
		So(err, ShouldBeNil)

		{
			f, err := selection.Fetch("name")
			So(err, ShouldBeNil)

			v, err := f.Value()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, "Adam")
		}

		{
			f, err := selection.Fetch("age")
			So(err, ShouldBeNil)

			v, err := f.Value()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 123)
		}
	})
}
