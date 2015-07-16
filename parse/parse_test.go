package parse

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNumeric(t *testing.T) {
	type numberTest struct {
		text    string
		isInt   bool
		isFloat bool
		int32
		float64
	}

	numberTests := []numberTest{
		// basics
		{"0", true, true, 0, 0},
		{"-0", true, true, 0, 0},
		{"73", true, true, 73, 73},
		{"073", true, true, 073, 073},
		{"-73", true, true, -73, -73},
		{"+73", true, true, 73, 73},
		{"100", true, true, 100, 100},
		{"-1.2", false, true, 0, -1.2},
	}

	Convey("Verify numbers parse correctly", t, func() {
		for _, test := range numberTests {
			n, err := New("test").newNumber(0, test.text, itemNumber)
			So(err, ShouldBeNil)

			if n.IsInt {
				So(test.isInt, ShouldBeTrue)
				So(n.Int32, ShouldEqual, test.int32)
			}

			if n.IsFloat {
				So(test.isFloat, ShouldBeTrue)
				So(n.Float64, ShouldEqual, test.float64)
			}
		}
	})
}

func TestQuery(t *testing.T) {
	type queryTest struct {
		text string
	}

	queryTests := []queryTest{
		{`{user(id:4) {name}}`},
	}

	Convey("Verify simple queries are parsed correctly", t, func() {
		for _, test := range queryTests {
			treeSet := make(map[string]*Tree)
			tree, err := New("test").Parse(test.text, treeSet)
			So(err, ShouldNotBeNil) // should err out for now
			So(tree, ShouldBeNil)
		}
	})
}
