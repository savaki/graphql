package restq

import (
	"io/ioutil"
	"net/http"
	"errors"

	"github.com/savaki/graphql"
	"github.com/savaki/graphql/provider/jsonq"
)

type Store struct {
	Client *http.Client
}

func New() *Store {
	return &Store{
		Client: http.DefaultClient,
	}
}

func (s *Store) Mutate(c *graphql.Context) (graphql.Field, error) {
	return nil, errors.New("#Mutate is not yet implemented")
}

func (s *Store) Query(c *graphql.Context) (graphql.Field, error) {
	switch c.Name {
	case "GET", "get":
		selection, err := s.get(c.Args[0].Value.(string))
		if err != nil {
			return nil, err
		}
		return field{selection: selection}, nil

	default:
		return nil, errors.New("Query only supports the GET method")
	}
}

func (s *Store) get(url string) (graphql.Selection, error) {
	resp, err := s.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return jsonq.New(data)
}
