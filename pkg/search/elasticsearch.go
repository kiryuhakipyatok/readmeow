package search

import (
	"context"
	"fmt"
	"readmeow/internal/config"
	"time"

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
		panic("failed to connect to elasticsearch: " + err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.PingTimeout))
	defer cancel()
	if _, err := client.Ping().Do(ctx); err != nil {
		panic("failed to ping elasticsearch: " + err.Error())
	}
	sc := &SearchClient{
		Client: client,
	}
	return sc
}
