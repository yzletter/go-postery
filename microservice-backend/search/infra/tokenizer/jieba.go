package tokenizer

import (
	"github.com/yanyiwu/gojieba"
)

type JiebaTokenizer struct {
}

func NewJiebaTokenizer() *JiebaTokenizer {
	return &JiebaTokenizer{}
}

func (tokenizer *JiebaTokenizer) Cut(text string) []string {
	// 初始化 gojieba
	x := gojieba.NewJieba()
	return x.CutForSearch(text, true)
}
