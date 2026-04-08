# bifrost-encoding-format-plugin

A [Bifrost](https://github.com/maximhq/bifrost) plugin that injects missing `encoding_format` parameters into embedding API requests.

> [!WARNING]
> This code in this project was **fully** generated using large language models (LLMs) in an approach called [vibe coding](https://en.wikipedia.org/wiki/Vibe_coding). The reason for taking this approach is that it allowed us to quickly resolve the problem described below, even though we had no experience with Go, which would be required when using a regular software development approach. You should **not rely** on this project for any critical tasks or use it for production setups of Bifrost.

## Problem

Some downstream LLM providers (e.g. older versions of LiteLLM) require the `encoding_format` parameter in embedding requests, but many upstream platforms like Open WebUI don't include it. This mismatch results in errors. This plugin bridges that gap by automatically adding the parameter when it's missing.

## How It Works

The plugin intercepts HTTP requests at the transport level before they reach the downstream provider. For any request targeting an embedding endpoint (path containing `embedding`):

1. If `encoding_format` is **missing** from the request body → injects the configured value
2. If `encoding_format` is **null** → replaces it with the configured value
3. If `encoding_format` is **already set** to a non-null value → leaves the request unchanged
4. Non-embedding requests are passed through without modification

## Installation

> [!IMPORTANT]
> Plugins require dynamic builds of Bifrost which are not enabled by default to keep Bifrost setup easier. You need to [build a dynamically linked Bifrost binary](https://docs.getbifrost.ai/plugins/building-dynamic-binary) to use this plugin.

You may install the plugin using two different approaches:

### Via the user interface

### Via the configuration file

Download the plugin and add it to your Bifrost configuration:

```yaml
plugins:
  - name: encoding-format
    enabled: true
    path: /path/to/encoding-format.so
    config:
      encoding_format: "float"  # or "base64"
```

If no `config` is provided, the plugin defaults to `float`.

## License

[MIT](LICENSE)
