package search

import (
	"context"
	"fmt"
	"readmeow/internal/config"
	"readmeow/pkg/logger"
	"time"

	es "github.com/elastic/go-elasticsearch/v9"
)

type SearchClient struct {
	Client *es.TypedClient
}

func MustConnect(cfg config.SearchConfig, l *logger.Logger) *SearchClient {
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
	for i := 0; i < 10; i++ {
		if _, err := client.Ping().Do(ctx); err != nil {
			break
		}
		l.Info("elasticsearch not ready yet (attempt %d/10), retrying...")
		time.Sleep(5 * time.Second)
	}
	sc := &SearchClient{
		Client: client,
	}
	return sc
}
