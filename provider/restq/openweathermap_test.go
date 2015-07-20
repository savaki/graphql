package restq

import (
	"testing"

	"bytes"

	"github.com/savaki/graphql"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOpenWeatherMap(t *testing.T) {
	Convey("Given the rest graphql handler", t, func() {
		type Result struct {
			Name    string
			Weather struct {
				Temp string
			}
		}

		query := `query city: GET("http://api.openweathermap.org/data/2.5/weather?lat=35&lon=139") {
			name
			weather: main {
				temp: temperature
			}
		}
		`

		buf := bytes.NewBuffer([]byte{})
		store := New()
		executor := gographql.New(store)
		err := executor.Handle(query, buf)
		So(err, ShouldNotBeNil)

		//		result := Result{}
		//		err = json.Unmarshal(buf.Bytes(), &result)
		//		So(err, ShouldBeNil)
	})
}
