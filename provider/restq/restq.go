package restq

import (
	"net/http"
	"io/ioutil"

	"github.com/docker/machine/drivers/vmwarevsphere/errors"
	"github.com/savaki/graphql"
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

func (s *Store) Mutate(name string, args ...gographql.Arg) (gographql.Field, error) {
	return nil, errors.New("#Mutate is not yet implemented")
}

func (s *Store) Query(name string, args ...gographql.Arg) (gographql.Field, error) {
	switch name {
	case "GET", "get":
		return s.get(args[0].Value.(string))
	default:
		return nil, errors.New("Query only supports the GET method")
	}
}

func (s *Store) get(url string) (gographql.Field, error) {
	resp, err := s.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return nil, errors.New("WIP")
}
