package mapq_test

import (
	"os"
	"testing"
	"bytes"

	"github.com/savaki/graphql"
	"github.com/savaki/graphql/provider/mapq"
	. "github.com/smartystreets/goconvey/convey"
)

func ExampleMap() {
	model := map[string]interface{}{"hello": "world"}
	store := mapq.New(model)
	graphql.New(store).Handle(`{hello}`, os.Stdout)
	// prints {"hello":"world"}
}

func TestHelloWorld(t *testing.T) {
	Convey("Given the hello world query", t, func() {
		model := map[string]interface{}{"hello": "world"}
		store := mapq.New(model)
		buf := bytes.NewBuffer([]byte{})
		err := graphql.New(store).Handle(`{hello}`, buf)

		So(err, ShouldBeNil)
		So(buf.String(), ShouldEqual, `{"hello":"world"}`)
	})
}
