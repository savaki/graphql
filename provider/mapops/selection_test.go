package mapops

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/savaki/gographql"
	. "github.com/smartystreets/goconvey/convey"
	"bitbucket.org/dataskoop/x/log"
)

func TestStore(t *testing.T) {
	Convey("Verify map store can handle gra", t, func() {
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

		buf := bytes.NewBuffer([]byte{})
		query := `query bill { friends }`
		err := gographql.New(store).Handle(query, buf)

		v := map[string]map[string][]string{}
		err = json.Unmarshal(buf.Bytes(), &v)
		So(err, ShouldBeNil)

		So(v, ShouldResemble, map[string]map[string][]string{
			"bill": map[string][]string{
				"friends": friends,
			},
		})
	})
}

func BenchmarkStore(b *testing.B) {
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
	buf := bytes.NewBuffer(make([]byte, 16384))

	for i := 0; i < b.N; i++ {
		buf.Reset()
		query := `query bill { friends }`
		err := gographql.New(store).Handle(query, buf)
		if err !=nil {
			log.Fatalln(err)
		}
	}
}