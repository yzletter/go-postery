//go:build !cgo

package tokenizer

import "strings"

type JiebaTokenizer struct {
}

func NewJiebaTokenizer() *JiebaTokenizer {
	return &JiebaTokenizer{}
}

func (tokenizer *JiebaTokenizer) Cut(text string) []string {
	words := strings.Fields(text)
	if len(words) == 0 && text != "" {
		return []string{text}
	}
	return words
}

func (tokenizer *JiebaTokenizer) Close() {
}
