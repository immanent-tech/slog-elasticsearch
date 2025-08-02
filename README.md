<h1 align="center">
  <a href="https://github.com/immanent-tech/slog-elasticsearch">
    <!-- Please provide path to your logo here -->
    <!-- <img src="docs/images/logo.svg" alt="Logo" width="100" height="100"> -->
  </a>
</h1>

<div align="center">
  slog-elasticsearch
  <br />
  <br />
  <br />
  <a href="https://github.com/immanent-tech/slog-elasticsearch/issues/new?assignees=&labels=bug&template=01_BUG_REPORT.md&title=bug%3A+">Report a Bug</a>
  ·
  <a href="https://github.com/immanent-tech/slog-elasticsearch/issues/new?assignees=&labels=enhancement&template=02_FEATURE_REQUEST.md&title=feat%3A+">Request a Feature</a>
  .
  <a href="https://github.com/immanent-tech/slog-elasticsearch/issues/new?assignees=&labels=question&template=04_SUPPORT_QUESTION.md&title=support%3A+">Ask a Question</a>
</div>

<div align="center">
<br />

[![Project license](https://img.shields.io/github/license/immanent-tech/slog-elasticsearch.svg?style=flat-square)](LICENSE)

[![Pull Requests welcome](https://img.shields.io/badge/PRs-welcome-ff69b4.svg?style=flat-square)](https://github.com/immanent-tech/slog-elasticsearch/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
[![code with love by immanent-tech](https://img.shields.io/badge/%3C%2F%3E%20with%20%E2%99%A5%20by-immanent-tech-ff1414.svg?style=flat-square)](https://github.com/immanent-tech)

</div>

<details open="open">
<summary>Table of Contents</summary>

- [About](#about)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
  - [Handler options](#handler-options)
  - [Example](#example)
- [Roadmap](#roadmap)
- [Support](#support)
- [Project assistance](#project-assistance)
- [Contributing](#contributing)
- [Authors \& contributors](#authors--contributors)
- [Security](#security)
- [License](#license)
- [Acknowledgements](#acknowledgements)

</details>

---

## About

An [slog](https://pkg.go.dev/log/slog) handler for logging to [Elasticsearch](https://elastic.co/elasticsearch).

Uses the the official [go-elasticsearch](https://github.com/elastic/go-elasticsearch) package under the hood. In
particular, it makes use of the bulk indexer helper provided in that package. Where appropriate, options for tuning the
bulk indexer are exposed by the handler.

## Getting Started

### Prerequisites

- Elasticsearch up and running.
- An index (or ideally, a [logs data
  stream](https://www.elastic.co/docs/manage-data/data-store/data-streams/logs-data-stream) created for the logs in
  Elasticsearch.

### Installation

```shell
go get github.com/immanent-tech/slog-elasticsearch/v2
```

## Usage


### Handler options

```go
type Option struct {
	// Log level (default: debug)
	Level slog.Leveler

	// Connection to Elasticsearch
	Conn *elasticsearch.Client
	// Index/alias to use for logging.
	Index string
	// Optional: The number of workers. Defaults to runtime.NumCPU().
	Numworkers int
	// Optional: The flush threshold in bytes. Defaults to 5MB.
	FlushBytes int
	// Optional: The flush threshold as duration. Defaults to 30sec.
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
```

Attributes will be injected in log payload.

Other global parameters:

```go
slogelasticsearch.SourceKey = "source"
slogelasticsearch.ContextKey = "extra"
slogelasticsearch.ErrorKeys = []string{"error", "err"}
```

### Example

```go
package main

import (
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
```
## Roadmap

See the [open issues](https://github.com/immanent-tech/slog-elasticsearch/issues) for a list of proposed features (and
known issues).

- [Top Feature Requests](https://github.com/immanent-tech/slog-elasticsearch/issues?q=label%3Aenhancement+is%3Aopen+sort%3Areactions-%2B1-desc) (Add your votes using the 👍 reaction)
- [Top Bugs](https://github.com/immanent-tech/slog-elasticsearch/issues?q=is%3Aissue+is%3Aopen+label%3Abug+sort%3Areactions-%2B1-desc) (Add your votes using the 👍 reaction)
- [Newest Bugs](https://github.com/immanent-tech/slog-elasticsearch/issues?q=is%3Aopen+is%3Aissue+label%3Abug)

## Support

Reach out to the maintainer at one of the following places:

- [GitHub issues](https://github.com/immanent-tech/slog-elasticsearch/issues/new?assignees=&labels=question&template=04_SUPPORT_QUESTION.md&title=support%3A+)

## Project assistance

If you want to say **thank you** or/and support active development of slog-elasticsearch:

- Add a [GitHub Star](https://github.com/immanent-tech/slog-elasticsearch) to the project.
- Post on social media about slog-elasticsearch.
- Write interesting articles about the project on [Dev.to](https://dev.to/), [Medium](https://medium.com/) or your
  personal blog.

Together, we can make slog-elasticsearch **better**!

## Contributing

First off, thanks for taking the time to contribute! Contributions are what make the open-source community such an
amazing place to learn, inspire, and create. Any contributions you make will benefit everybody else and are **greatly
appreciated**.

Please read [our contribution guidelines](docs/CONTRIBUTING.md), and thank you for being involved!

## Authors & contributors

The original setup of this repository is by [joshuar](https://github.com/joshuar).

For a full list of all authors and contributors, see [the contributors page](https://github.com/immanent-tech/slog-elasticsearch/contributors).

## Security

slog-elasticsearch follows good practices of security, but 100% security cannot be assured.
slog-elasticsearch is provided **"as is"** without any **warranty**. Use at your own risk.

_For more information and to report security issues, please refer to our [security documentation](docs/SECURITY.md)._

## License

This project is licensed under the **MIT license**.

See [LICENSE](LICENSE) for more information.

## Acknowledgements

slog-elasticsearch was original forked from and based on the code in
[slog-logstash](https://github.com/samber/slog-logstash). Many thanks go to the authors and contributors of that project
for a great basis for slog-elasticsearch.
