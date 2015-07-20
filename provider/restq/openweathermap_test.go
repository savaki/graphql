package restq

import (
	"bytes"
	"testing"

	"github.com/savaki/graphql"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOpenWeatherMap(t *testing.T) {
	Convey("Given the rest graphql handler", t, func() {
		type Result struct {
			City struct {
				Name    string
				Weather struct {
					Temperature float32
				}
			}
		}

		query := `query city: GET(url:"http://api.openweathermap.org/data/2.5/weather?lat=35&lon=139") {
			name
			weather: main {
				temperature: temp
			}
		}`

		buf := bytes.NewBuffer([]byte{})
		store := New()
		executor := graphql.New(store)

		So(query, ShouldNotBeNil)
		So(buf, ShouldNotBeNil)
		So(store, ShouldNotBeNil)
		So(executor, ShouldNotBeNil)

		//		err := executor.Handle(query, buf)
		//		So(err, ShouldBeNil)
		//
		//		result := Result{}
		//		err = json.Unmarshal(buf.Bytes(), &result)
		//		So(err, ShouldBeNil)
		//
		//		So(result.City.Name, ShouldEqual, "Shuzenji")
		//		So(result.City.Weather.Temperature, ShouldBeGreaterThan, 0.0)
	})
}
