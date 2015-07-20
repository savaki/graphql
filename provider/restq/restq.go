package restq

import (
	"io/ioutil"
	"net/http"

	"github.com/docker/machine/drivers/vmwarevsphere/errors"
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

func WithClient(store *Store, client *http.Client) *Store {
	return &Store{
		Client: client,
	}
}

func (s *Store) Mutate(c *graphql.Context) (graphql.Selection, error) {
	return nil, errors.New("#Mutate is not yet implemented")
}

func (s *Store) Query(c *graphql.Context) (graphql.Selection, error) {
	switch c.Name {
	case "GET", "get":
		return s.get(c.Args[0].Value.(string))

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
