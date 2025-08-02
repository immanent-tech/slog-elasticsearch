package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	slogelasticsearch "github.com/immanent-tech/slog-elasticsearch/v2"
)

func main() {
	// Create the Elasticsearch client
	//
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Create a logger using the slog-elasticsearch handler.
	//
	logger := slog.New(slogelasticsearch.Option{
		Level: slog.LevelDebug,
		Conn:  es,
		Index: "test-logs",
	}.NewElasticsearchHandler(context.Background()))

	// Use the logger.
	//
	logger = logger.With("release", "v1.0.0")

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
			),
		).
		With("environment", "dev").
		With("error", fmt.Errorf("an error")).
		Error("A message")
}
