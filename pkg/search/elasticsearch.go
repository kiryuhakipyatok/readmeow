package search

import (
	"readmeow/internal/config"

	es "github.com/elastic/go-elasticsearch/v9"
)

type SearchClient struct {
	Client *es.TypedClient
}

func Connect(cfg *config.SearchConfig) *SearchClient {
	client, err := es.NewTypedClient(es.Config{
		Addresses: []string{cfg.Host + "" + cfg.Port},
		Username:  cfg.User,
		Password:  cfg.Password,
	})
	if err != nil {
		panic("failed to connect to elasticsearch")
	}
	sc := &SearchClient{
		Client: client,
	}
	return sc
}
