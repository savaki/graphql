package restq_test

import (
	"os"

	"github.com/savaki/graphql/provider/restq"
	"github.com/savaki/graphql"
)

func ExampleGet() {
	query := `query city: GET("http://api.openweathermap.org/data/2.5/weather?q=London") {
		name
		weather: main {
			temperature: temp
		}
	}`

	store := restq.New()
	graphql.New(store).Handle(query, os.Stdout)
}
