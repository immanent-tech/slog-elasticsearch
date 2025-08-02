package slogelasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
	slogcommon "github.com/samber/slog-common"
)

type Option struct {
	// Log level (default: debug)
	Level slog.Leveler

	// Connection to Elasticsearch
	Conn *elasticsearch.Client
	// Index/alias to use for logging.
	Index string
	// Optional: The number of workers for the bulk indexer. Defaults to runtime.NumCPU().
	Numworkers int
	// Optional: The flush threshold for the bulk indexer in bytes. Defaults to 5MB.
	FlushBytes int
	// Optional: The flush threshold as duration for the bulk indexer. Defaults to 30sec.
	FlushInterval time.Duration

	// Optional: customize json payload builder
	Converter Converter
	// Optional: custom marshaler
	Marshaler func(v any) ([]byte, error)
	// Optional: fetch attributes from context
	AttrFromContext []func(ctx context.Context) []slog.Attr

	// Optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o Option) NewElasticsearchHandler(ctx context.Context) slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelDebug
	}

	if o.Conn == nil {
		panic("missing elasticsearch connection")
	}

	config := esutil.BulkIndexerConfig{
		Index:  o.Index, // The default index name
		Client: o.Conn,  // The Elasticsearch client
	}
	// Set optional parameters if set.
	if o.Numworkers != 0 {
		config.NumWorkers = o.Numworkers
	}
	if o.FlushBytes != 0 {
		config.FlushBytes = o.FlushBytes
	}
	if o.FlushInterval != 0 {
		config.FlushInterval = o.FlushInterval
	}

	// Create a bulk indexer.
	indexer, err := esutil.NewBulkIndexer(config)
	if err != nil {
		panic(fmt.Sprintf("error creating the indexer: %v", err))
	}

	if o.Converter == nil {
		o.Converter = DefaultConverter
	}

	if o.Marshaler == nil {
		o.Marshaler = json.Marshal
	}

	if o.AttrFromContext == nil {
		o.AttrFromContext = []func(ctx context.Context) []slog.Attr{}
	}

	go func() {
		defer indexer.Close(ctx)
		<-ctx.Done()
	}()

	return &ElasticsearchHandler{
		option:  o,
		attrs:   []slog.Attr{},
		groups:  []string{},
		indexer: indexer,
	}
}

var _ slog.Handler = (*ElasticsearchHandler)(nil)

type ElasticsearchHandler struct {
	option  Option
	attrs   []slog.Attr
	groups  []string
	indexer esutil.BulkIndexer
}

func (h *ElasticsearchHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *ElasticsearchHandler) Handle(ctx context.Context, record slog.Record) error {
	fromContext := slogcommon.ContextExtractor(ctx, h.option.AttrFromContext)
	message := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, append(h.attrs, fromContext...), h.groups, &record)

	data, err := h.option.Marshaler(message)
	if err != nil {
		return fmt.Errorf("unable to send log message: %w", err)
	}

	go func() {
		err = h.indexer.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action: "index",
				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(data),
			},
		)
		if err != nil {
			slog.Error("Unexpected error.",
				slog.Any("error", err),
			)
		}
	}()

	return nil
}

func (h *ElasticsearchHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ElasticsearchHandler{
		option: h.option,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *ElasticsearchHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &ElasticsearchHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}
