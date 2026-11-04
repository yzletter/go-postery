package llm

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yzletter/go-postery/backend/conf"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestQwenEmbedderUsesConfiguredEndpoint(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://embedding.example.test/v1/embeddings" {
			t.Errorf("unexpected request URL: %s", req.URL)
		}
		if req.Header.Get("Authorization") != "Bearer embedding-test-key" {
			t.Error("embedding API key was not used")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":[{"embedding":[3,4],"index":0}],"model":"test-model"}`))}, nil
	})}
	got := newQwenEmbedder(context.Background(), conf.QwenConfig{
		LLMBaseURL: "https://llm.example.test/v1", EmbedderBaseURL: "https://embedding.example.test/v1",
		EmbedderAPIKey: "embedding-test-key", EmbedderModel: "test-model",
	}, client)
	if got == nil {
		t.Fatal("failed to initialize embedder")
	}
	vectors, err := got.EmbedStrings(context.Background(), []string{"test"})
	if err != nil || len(vectors) != 1 || len(vectors[0]) != 2 || vectors[0][0] != 0.6 || vectors[0][1] != 0.8 {
		t.Fatalf("unexpected embedding result: %v, %v", vectors, err)
	}
}
