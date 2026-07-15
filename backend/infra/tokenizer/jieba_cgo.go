//go:build cgo

package tokenizer

import "github.com/yanyiwu/gojieba"

type JiebaTokenizer struct {
}

func NewJiebaTokenizer() *JiebaTokenizer {
	return &JiebaTokenizer{}
}

func (tokenizer *JiebaTokenizer) CutSearch(text string) []string {
	x := gojieba.NewJieba()
	defer x.Free()
	return x.CutForSearch(text, true)
}

func (tokenizer *JiebaTokenizer) Close() {
}
