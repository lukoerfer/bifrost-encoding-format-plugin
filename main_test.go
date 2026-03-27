package main

import (
	"context"
	"testing"
	"time"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/tidwall/gjson"
)

func newTestContext() *schemas.BifrostContext {
	return schemas.NewBifrostContext(context.Background(), time.Now().Add(time.Minute))
}

func TestGetName(t *testing.T) {
	name := GetName()
	if name != "encoding-format" {
		t.Errorf("expected plugin name 'encoding-format', got '%s'", name)
	}
}

func TestInitDefault(t *testing.T) {
	activeConfig = pluginConfig{}
	if err := Init(nil); err != nil {
		t.Fatalf("Init(nil) returned error: %v", err)
	}
	if activeConfig.EncodingFormat != "float" {
		t.Errorf("expected default encoding_format 'float', got '%s'", activeConfig.EncodingFormat)
	}
}

func TestInitWithFloatConfig(t *testing.T) {
	activeConfig = pluginConfig{}
	cfg := map[string]interface{}{"encoding_format": "float"}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init with float config returned error: %v", err)
	}
	if activeConfig.EncodingFormat != "float" {
		t.Errorf("expected encoding_format 'float', got '%s'", activeConfig.EncodingFormat)
	}
}

func TestInitWithBase64Config(t *testing.T) {
	activeConfig = pluginConfig{}
	cfg := map[string]interface{}{"encoding_format": "base64"}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init with base64 config returned error: %v", err)
	}
	if activeConfig.EncodingFormat != "base64" {
		t.Errorf("expected encoding_format 'base64', got '%s'", activeConfig.EncodingFormat)
	}
}

func TestInitWithUnsupportedFormat(t *testing.T) {
	activeConfig = pluginConfig{}
	cfg := map[string]interface{}{"encoding_format": "unsupported"}
	err := Init(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported encoding_format, got nil")
	}
}

func TestInitWithEmptyConfig(t *testing.T) {
	activeConfig = pluginConfig{}
	cfg := map[string]interface{}{}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init with empty config returned error: %v", err)
	}
	if activeConfig.EncodingFormat != "float" {
		t.Errorf("expected default encoding_format 'float', got '%s'", activeConfig.EncodingFormat)
	}
}

func TestIsEmbeddingRequest(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"standard embeddings path", "/v1/embeddings", true},
		{"embedding singular", "/v1/embedding", true},
		{"uppercase path", "/V1/EMBEDDINGS", true},
		{"mixed case", "/v1/Embeddings", true},
		{"nested path", "/api/v1/embeddings/create", true},
		{"chat path", "/v1/chat/completions", false},
		{"completions path", "/v1/completions", false},
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &schemas.HTTPRequest{Path: tt.path}
			result := isEmbeddingRequest(req)
			if result != tt.expected {
				t.Errorf("isEmbeddingRequest(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestHTTPTransportPreHookInjectsEncodingFormat(t *testing.T) {
	activeConfig = pluginConfig{EncodingFormat: "float"}
	ctx := newTestContext()

	req := &schemas.HTTPRequest{
		Path: "/v1/embeddings",
		Body: []byte(`{"model":"text-embedding-ada-002","input":"hello"}`),
	}

	resp, err := HTTPTransportPreHook(ctx, req)
	if err != nil {
		t.Fatalf("HTTPTransportPreHook returned error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response (continue processing)")
	}

	result := gjson.GetBytes(req.Body, "encoding_format")
	if !result.Exists() {
		t.Fatal("encoding_format was not injected into request body")
	}
	if result.String() != "float" {
		t.Errorf("expected encoding_format 'float', got '%s'", result.String())
	}
}

func TestHTTPTransportPreHookInjectsBase64(t *testing.T) {
	activeConfig = pluginConfig{EncodingFormat: "base64"}
	ctx := newTestContext()

	req := &schemas.HTTPRequest{
		Path: "/v1/embeddings",
		Body: []byte(`{"model":"text-embedding-ada-002","input":"hello"}`),
	}

	resp, err := HTTPTransportPreHook(ctx, req)
	if err != nil {
		t.Fatalf("HTTPTransportPreHook returned error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}

	result := gjson.GetBytes(req.Body, "encoding_format")
	if result.String() != "base64" {
		t.Errorf("expected encoding_format 'base64', got '%s'", result.String())
	}
}

func TestHTTPTransportPreHookSkipsExistingFormat(t *testing.T) {
	activeConfig = pluginConfig{EncodingFormat: "float"}
	ctx := newTestContext()

	req := &schemas.HTTPRequest{
		Path: "/v1/embeddings",
		Body: []byte(`{"model":"text-embedding-ada-002","input":"hello","encoding_format":"base64"}`),
	}

	originalBody := string(req.Body)

	resp, err := HTTPTransportPreHook(ctx, req)
	if err != nil {
		t.Fatalf("HTTPTransportPreHook returned error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}

	if string(req.Body) != originalBody {
		t.Error("request body was modified when encoding_format was already set")
	}

	result := gjson.GetBytes(req.Body, "encoding_format")
	if result.String() != "base64" {
		t.Errorf("expected existing encoding_format 'base64' to be preserved, got '%s'", result.String())
	}
}

func TestHTTPTransportPreHookInjectsWhenNull(t *testing.T) {
	activeConfig = pluginConfig{EncodingFormat: "float"}
	ctx := newTestContext()

	req := &schemas.HTTPRequest{
		Path: "/v1/embeddings",
		Body: []byte(`{"model":"text-embedding-ada-002","input":"hello","encoding_format":null}`),
	}

	resp, err := HTTPTransportPreHook(ctx, req)
	if err != nil {
		t.Fatalf("HTTPTransportPreHook returned error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}

	result := gjson.GetBytes(req.Body, "encoding_format")
	if result.String() != "float" {
		t.Errorf("expected encoding_format 'float' after null replacement, got '%s'", result.String())
	}
}

func TestHTTPTransportPreHookSkipsNonEmbeddingRequests(t *testing.T) {
	activeConfig = pluginConfig{EncodingFormat: "float"}
	ctx := newTestContext()

	req := &schemas.HTTPRequest{
		Path: "/v1/chat/completions",
		Body: []byte(`{"model":"gpt-4","messages":[{"role":"user","content":"hello"}]}`),
	}

	originalBody := string(req.Body)

	resp, err := HTTPTransportPreHook(ctx, req)
	if err != nil {
		t.Fatalf("HTTPTransportPreHook returned error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}

	if string(req.Body) != originalBody {
		t.Error("non-embedding request body was modified")
	}
}

func TestHTTPTransportPreHookSkipsEmptyBody(t *testing.T) {
	activeConfig = pluginConfig{EncodingFormat: "float"}
	ctx := newTestContext()

	req := &schemas.HTTPRequest{
		Path: "/v1/embeddings",
		Body: []byte{},
	}

	resp, err := HTTPTransportPreHook(ctx, req)
	if err != nil {
		t.Fatalf("HTTPTransportPreHook returned error: %v", err)
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}
}

func TestCleanup(t *testing.T) {
	if err := Cleanup(); err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
}
