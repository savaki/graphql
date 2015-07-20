package jsonq

import (
	"testing"

	"github.com/savaki/graphql"
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

		field, err := store.Fetch(&graphql.Context{Name: "a"})
		So(err, ShouldBeNil)

		selection, err := field.Selection()
		So(err, ShouldBeNil)

		{
			f, err := selection.Query(&graphql.Context{Name: "name"})
			So(err, ShouldBeNil)

			v, err := f.Value()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, "Adam")
		}

		{
			f, err := selection.Query(&graphql.Context{Name: "age"})
			So(err, ShouldBeNil)

			v, err := f.Value()
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 123)
		}
	})
}
