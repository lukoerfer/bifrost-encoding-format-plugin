# bifrost-encoding-format-plugin

A [Bifrost](https://github.com/maximhq/bifrost) plugin that injects missing `encoding_format` parameters into embedding API requests.

## Problem

Some downstream LLM providers (e.g. older versions of LiteLLM) require the `encoding_format` parameter in embedding requests, but many upstream platforms like Open WebUI don't include it. This mismatch results in errors. This plugin bridges that gap by automatically adding the parameter when it's missing.

## How It Works

The plugin intercepts HTTP requests at the transport level before they reach the downstream provider. For any request targeting an embedding endpoint (path containing `embedding`):

1. If `encoding_format` is **missing** from the request body → injects the configured value
2. If `encoding_format` is **null** → replaces it with the configured value
3. If `encoding_format` is **already set** to a non-null value → leaves the request unchanged
4. Non-embedding requests are passed through without modification

## Configuration

Add the plugin to your Bifrost configuration:

```yaml
plugins:
  - name: encoding-format
    enabled: true
    path: /path/to/encoding-format.so
    config:
      encoding_format: "float"  # or "base64"
```

### Supported Values

| Value    | Description                        |
|----------|------------------------------------|
| `float`  | Return embeddings as float arrays (default) |
| `base64` | Return embeddings as base64-encoded strings |

If no configuration is provided, the plugin defaults to `float`.

## Building

### Prerequisites

- Go 1.26.1+
- CGO enabled (required for Go plugins)

### Build locally

```bash
make build
```

### Build for Linux AMD64 (cross-compilation via Docker)

```bash
make build GOOS=linux GOARCH=amd64
```

### Install to Bifrost plugins directory

```bash
make install
```

## GitHub Actions

The repository includes a GitHub Actions workflow that:

- **On push/PR to main**: Builds and tests the plugin on Linux AMD64
- **On tag push (`v*`)**: Builds the plugin and publishes it as a GitHub Release asset

## Development

```bash
# Download dependencies
make deps

# Build for development (no optimizations)
make dev

# Run tests
make test
```

## License

[MIT](LICENSE)
