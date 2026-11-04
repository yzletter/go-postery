package conf

import (
	"context"
	"testing"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

type qwenTestReader map[string]string

func (r qwenTestReader) Get(_ context.Context, key string, _ ...etcdv3.OpOption) (*etcdv3.GetResponse, error) {
	resp := &etcdv3.GetResponse{}
	if value, ok := r[key]; ok {
		resp.Kvs = []*mvccpb.KeyValue{{Key: []byte(key), Value: []byte(value)}}
	}
	return resp, nil
}

func TestLoadQwenConfigSeparateAPIKeys(t *testing.T) {
	const prefix = "test/interview/"
	reader := qwenTestReader{
		prefix + "qwen/llm_api_key":       "llm-test-key",
		prefix + "qwen/embedder_api_key":  "embedding-test-key",
		prefix + "qwen/api_key":           "legacy-test-key",
		prefix + "qwen/llm_base_url":      "https://llm.example.test/v1",
		prefix + "qwen/embedder_base_url": "https://embedding.example.test/v1",
		prefix + "qwen/base_url":          "https://old.example.test/v1",
	}
	got := loadQwenConfig(context.Background(), reader, prefix)
	if got.LLMAPIKey != "llm-test-key" || got.EmbedderAPIKey != "embedding-test-key" {
		t.Fatal("LLM and embedding API keys were not loaded independently")
	}
	if got.LLMBaseURL != "https://llm.example.test/v1" || got.EmbedderBaseURL != "https://embedding.example.test/v1" {
		t.Fatal("LLM and embedding base URLs were not loaded independently")
	}
	for _, key := range []string{"qwen/llm_api_key", "qwen/embedder_api_key", "qwen/llm_base_url", "qwen/embedder_base_url"} {
		if _, ok := watchKeys[prefix+key]; !ok {
			t.Errorf("%s was not registered for change notification", key)
		}
		delete(watchKeys, prefix+key)
	}

	got = loadQwenConfig(context.Background(), qwenTestReader{prefix + "qwen/api_key": "legacy-test-key", prefix + "qwen/base_url": "https://old.example.test/v1"}, prefix)
	if got.LLMAPIKey != "" || got.EmbedderAPIKey != "" || got.LLMBaseURL != "" || got.EmbedderBaseURL != "" {
		t.Fatal("legacy shared configuration must not replace the separate keys")
	}
}
