package slogelasticsearch

import (
	"log/slog"
	"os"

	slogcommon "github.com/samber/slog-common"
)

var (
	SourceKey  = "source"
	ContextKey = "extra"
	ErrorKeys  = []string{"error", "err"}
)

type Converter func(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) map[string]any

func DefaultConverter(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) map[string]any {
	// aggregate all attributes
	attrs := slogcommon.AppendRecordAttrsToAttrs(loggerAttr, groups, record)

	// developer formatters
	if addSource {
		attrs = append(attrs, slogcommon.Source(SourceKey, record))
	}
	attrs = slogcommon.ReplaceAttrs(replaceAttr, []string{}, attrs...)
	attrs = slogcommon.RemoveEmptyAttrs(attrs)

	// handler formatter
	log := map[string]any{
		"@timestamp": record.Time.UTC(),
		"logger": map[string]any{
			"name":    name,
			"version": version,
		},
		"level":   record.Level.String(),
		"message": record.Message,
	}

	hostname, err := os.Hostname()
	if err == nil {
		log["host"] = map[string]any{
			"name": hostname,
		}
	}

	extra := slogcommon.AttrsToMap(attrs...)

	for _, errorKey := range ErrorKeys {
		if err, ok := extra[errorKey]; ok {
			if e, ok := err.(error); ok {
				log[errorKey] = slogcommon.FormatError(e)
				delete(extra, errorKey)
				break
			}
		}
	}

	log[ContextKey] = extra

	return log
}
