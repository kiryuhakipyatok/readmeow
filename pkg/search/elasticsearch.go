package search

import (
	"context"
	"fmt"
	"readmeow/internal/config"

	es "github.com/elastic/go-elasticsearch/v9"
)

type SearchClient struct {
	Client *es.TypedClient
}

func MustConnect(cfg config.SearchConfig) *SearchClient {
	client, err := es.NewTypedClient(es.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)},
		Username:  cfg.User,
		Password:  cfg.Password,
	})
	if err != nil {
		panic(fmt.Errorf("failed to connect to elasticsearch: %w", err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()
	if _, err := client.Ping().Do(ctx); err != nil {
		panic(fmt.Errorf("failed to ping elasticsearch: %w", err))
	}
	sc := &SearchClient{
		Client: client,
	}
	return sc
}
