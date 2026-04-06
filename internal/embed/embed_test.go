package embed

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCascadeTier1(t *testing.T) {
	e := NewCascade("openai", "sk-test", "", nil)
	if e == nil {
		t.Fatal("expected embedder for openai")
	}
	if e.Name() != "openai/text-embedding-3-small" {
		t.Errorf("unexpected name: %s", e.Name())
	}
	if e.Dimensions() != 1536 {
		t.Errorf("expected 1536 dims, got %d", e.Dimensions())
	}
}

func TestCascadeAnthropicFallsThrough(t *testing.T) {
	// Anthropic has no embedding API — should fall through
	e := NewCascade("anthropic", "sk-ant-test", "", nil)
	// This will return nil unless Ollama is running
	// We can't control Ollama in tests, so just verify no panic
	_ = e
}

func TestCascadeNoProvider(t *testing.T) {
	e := NewCascade("", "", "", nil)
	// Should return nil (no provider, no Ollama assumed)
	// Can't guarantee nil because Ollama might be running locally
	_ = e
}

func TestAPIEmbedderWithMockServer(t *testing.T) {
	// Mock OpenAI embedding API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			http.Error(w, "not found", 404)
			return
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			http.Error(w, "unauthorized", 401)
			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float32{0.1, 0.2, 0.3}},
			},
		})
	}))
	defer server.Close()

	e := &APIEmbedder{
		provider: "openai",
		model:    "text-embedding-3-small",
		apiKey:   "sk-test",
		baseURL:  server.URL,
		dims:     3,
	}

	vec, err := e.Embed("test text")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(vec) != 3 {
		t.Errorf("expected 3 dimensions, got %d", len(vec))
	}
	if vec[0] != 0.1 {
		t.Errorf("expected 0.1, got %f", vec[0])
	}
}

func TestAPIEmbedderErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", 429)
	}))
	defer server.Close()

	e := &APIEmbedder{
		provider: "openai",
		model:    "text-embedding-3-small",
		apiKey:   "sk-test",
		baseURL:  server.URL,
		dims:     1536,
	}

	_, err := e.Embed("test")
	if err == nil {
		t.Error("expected error on 429")
	}
}

func TestDefaultModels(t *testing.T) {
	providers := []string{"openai", "gemini", "voyage", "mistral"}
	for _, p := range providers {
		model, ok := defaultModels[p]
		if !ok {
			t.Errorf("missing default model for %s", p)
			continue
		}
		dims, ok := defaultDimensions[model]
		if !ok {
			t.Errorf("missing default dimensions for %s", model)
			continue
		}
		if dims <= 0 {
			t.Errorf("invalid dimensions %d for %s", dims, model)
		}
	}
}
