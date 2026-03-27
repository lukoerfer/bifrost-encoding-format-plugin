package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// supportedFormats lists the valid encoding_format values.
var supportedFormats = map[string]bool{
	"float":  true,
	"base64": true,
}

// pluginConfig holds the parsed plugin configuration.
type pluginConfig struct {
	EncodingFormat string `json:"encoding_format"`
}

// activeConfig stores the plugin configuration after initialization.
var activeConfig pluginConfig

// Init parses the plugin configuration and validates the encoding_format value.
// The config is expected to be a JSON object with an "encoding_format" field.
// If no config is provided, the default encoding_format "float" is used.
func Init(config any) error {
	activeConfig.EncodingFormat = "float"

	if config == nil {
		fmt.Printf("[encoding-format] No config provided, using default encoding_format: %s\n", activeConfig.EncodingFormat)
		return nil
	}

	cfgBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("[encoding-format] failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(cfgBytes, &activeConfig); err != nil {
		return fmt.Errorf("[encoding-format] failed to parse config: %w", err)
	}

	if activeConfig.EncodingFormat == "" {
		activeConfig.EncodingFormat = "float"
	}

	if !supportedFormats[activeConfig.EncodingFormat] {
		return fmt.Errorf("[encoding-format] unsupported encoding_format: %s (supported: float, base64)", activeConfig.EncodingFormat)
	}

	fmt.Printf("[encoding-format] Initialized with encoding_format: %s\n", activeConfig.EncodingFormat)
	return nil
}

// GetName returns the system identifier for this plugin (required).
func GetName() string {
	return "encoding-format"
}

// HTTPTransportPreHook intercepts HTTP requests before they are sent to the
// downstream provider. For embedding requests that are missing the
// encoding_format parameter (or have it set to null), it injects the
// configured encoding_format value into the request body.
func HTTPTransportPreHook(ctx *schemas.BifrostContext, req *schemas.HTTPRequest) (*schemas.HTTPResponse, error) {
	if !isEmbeddingRequest(req) {
		return nil, nil
	}

	if len(req.Body) == 0 {
		return nil, nil
	}

	result := gjson.GetBytes(req.Body, "encoding_format")
	if result.Exists() && result.Type != gjson.Null {
		return nil, nil
	}

	newBody, err := sjson.SetBytes(req.Body, "encoding_format", activeConfig.EncodingFormat)
	if err != nil {
		return nil, fmt.Errorf("[encoding-format] failed to set encoding_format in request body: %w", err)
	}

	req.Body = newBody
	return nil, nil
}

// isEmbeddingRequest checks whether the HTTP request targets an embedding endpoint.
func isEmbeddingRequest(req *schemas.HTTPRequest) bool {
	return strings.Contains(strings.ToLower(req.Path), "embedding")
}

// Cleanup performs any necessary teardown when the plugin is unloaded.
func Cleanup() error {
	return nil
}
