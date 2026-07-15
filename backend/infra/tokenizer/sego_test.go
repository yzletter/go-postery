package tokenizer

import "testing"

func TestSegoTokenizerCutLoadsDictionary(t *testing.T) {
	tokenizer := NewSegoTokenizer()
	defer tokenizer.Close()

	segments := tokenizer.CutSearch("中华人民共和国中央人民政府")
	if len(segments) == 0 {
		t.Fatal("expected non-empty segments")
	}
}

func TestSegoTokenizerZeroValuePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for zero-value SegoTokenizer")
		}
	}()

	var tokenizer SegoTokenizer
	_ = tokenizer.CutSearch("搜索")
}
