package slogelasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
	slogcommon "github.com/samber/slog-common"
)

type Option struct {
	// log level (default: debug)
	Level slog.Leveler

	// Connection to Elasticsearch
	Conn *elasticsearch.Client
	// Index/alias to use for logging.
	Index string
	// The number of workers. Defaults to runtime.NumCPU().
	Numworkers int
	// The flush threshold in bytes. Defaults to 5MB.
	FlushBytes int
	// The flush threshold as duration. Defaults to 30sec.
	FlushInterval time.Duration

	// optional: customize json payload builder
	Converter Converter
	// optional: custom marshaler
	Marshaler func(v any) ([]byte, error)
	// optional: fetch attributes from context
	AttrFromContext []func(ctx context.Context) []slog.Attr

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o Option) NewElasticsearchHandler() slog.Handler {
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
		panic(fmt.Errorf("error creating the indexer: %s", err))
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
		return err
	}

	go func() {
		err = h.indexer.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				// DocumentID is the (optional) document ID
				// DocumentID: strconv.Itoa(a.ID),

				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(data),

				// OnSuccess is called for each successful operation
				// OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				// 	atomic.AddUint64(&countSuccessful, 1)
				// },

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			slog.Error("Unexpected error.",
				slog.Any("error", err),
			)
		}
		// _, _ = h.option.Conn.Write(append(bytes, byte('\n')))
	}()

	return err
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
