# kraken-go

[![ci](https://github.com/kraken-io/kraken-go/actions/workflows/ci.yml/badge.svg)](https://github.com/kraken-io/kraken-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/kraken-io/kraken-go.svg)](https://pkg.go.dev/github.com/kraken-io/kraken-go)

The official Go client for the [Kraken.io](https://kraken.io) image API.

Requires Go 1.21 or newer. Tested against 1.21 through 1.24.

## Install

```bash
go get github.com/kraken-io/kraken-go
```

## Usage

Optimize an image by URL:

```go
package main

import (
	"log"

	"github.com/kraken-io/kraken-go"
)

func main() {
	kr, err := kraken.New("your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	data, err := kr.URL(map[string]interface{}{
		"wait":  true,
		"lossy": true,
		"url":   "https://example.com/file.jpg",
	})
	if err != nil {
		log.Fatal(err)
	}

	if data["success"] != true {
		log.Fatal("failed: ", data["message"])
	}
	log.Println("optimized:", data["kraked_url"])
}
```

Upload a file from disk:

```go
data, err := kr.Upload(map[string]interface{}{
	"wait":  true,
	"lossy": true,
}, "./photo.jpg")
```

Upload from any `io.Reader` — an HTTP body, an S3 object, a buffer:

```go
data, err := kr.UploadReader(map[string]interface{}{"wait": true}, r, "photo.jpg")
```

## Parameters

`params` is passed to the API as-is, so **every API feature is available** —
resizing, format conversion, metadata, WebP and AVIF output, and the cloud
storage destinations. See the [API documentation](https://kraken.io/docs) for
the full list.

```go
data, err := kr.URL(map[string]interface{}{
	"wait":  true,
	"lossy": true,
	"url":   "https://example.com/file.jpg",
	"resize": map[string]interface{}{
		"width":    800,
		"height":   600,
		"strategy": "fit",
	},
	"convert": map[string]interface{}{
		"format": "webp",
	},
})
```

`auth` is added for you from the credentials given to `New`; do not set it.

## Errors

Two kinds:

| | |
|---|---|
| `error` from the call | transport failure, or a response that was not JSON |
| `data["success"] == false` | the request reached the API and it declined |

A non-JSON response comes back as `*ResponseError`, which carries the
`*http.Response` so the status and body can be inspected:

```go
if respErr, ok := err.(*kraken.ResponseError); ok {
	log.Println("status:", respErr.Resp.StatusCode)
}
```

`New` returns `ErrNoCred` if either credential is empty.

## Custom HTTP client

`HTTPClient` is exported, so timeouts, proxies and transports are yours to set:

```go
kr.HTTPClient = &http.Client{Timeout: 30 * time.Second}
```

## Development

```bash
go test ./...          # no network and no credentials required
go vet ./...
gofmt -l .
```

Tests run against a local `httptest` server. The examples in `example_test.go`
are compiled by CI, so the usage documented here cannot drift from the code.

## License

See [LICENSE](LICENSE).
