package tokenizer

import (
	"github.com/huichen/sego"
)

type SegoTokenizer struct {
	segmenter sego.Segmenter
}

func NewSegoTokenizer() *SegoTokenizer {
	return &SegoTokenizer{
		segmenter: sego.Segmenter{},
	}
}

func (tokenizer *SegoTokenizer) Cut(text string) []string {
	segments := tokenizer.segmenter.Segment([]byte(text))
	res := sego.SegmentsToSlice(segments, true)
	return res
}

func (tokenizer *SegoTokenizer) Close() {
}
