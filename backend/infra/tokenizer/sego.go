package tokenizer

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/huichen/sego"
)

//go:embed data/dictionary.txt.gz
var defaultSegoDictionaryGzip []byte

var (
	defaultDictionaryOnce sync.Once
	defaultDictionaryPath string
	defaultDictionaryErr  error
)

type SegoTokenizer struct {
	segmenter sego.Segmenter
}

func NewSegoTokenizer() *SegoTokenizer {
	dictionaryPath, err := defaultSegoDictionaryFile()
	if err != nil {
		panic(fmt.Errorf("prepare sego dictionary: %w", err))
	}

	tokenizer := &SegoTokenizer{}
	tokenizer.segmenter.LoadDictionary(dictionaryPath)
	return tokenizer
}

func (tokenizer *SegoTokenizer) CutSearch(text string) []string {
	if tokenizer == nil || tokenizer.segmenter.Dictionary() == nil {
		panic("sego tokenizer is not initialized; use NewSegoTokenizer")
	}

	segments := tokenizer.segmenter.Segment([]byte(text))
	res := sego.SegmentsToSlice(segments, true)
	return res
}

func (tokenizer *SegoTokenizer) Cut(text string) []string {
	if tokenizer == nil || tokenizer.segmenter.Dictionary() == nil {
		panic("sego tokenizer is not initialized; use NewSegoTokenizer")
	}

	segments := tokenizer.segmenter.Segment([]byte(text))
	res := sego.SegmentsToSlice(segments, false)
	return res
}

func (tokenizer *SegoTokenizer) Close() {
	if tokenizer != nil {
		tokenizer.segmenter.Close()
		tokenizer.segmenter = sego.Segmenter{}
	}
}

func defaultSegoDictionaryFile() (string, error) {
	defaultDictionaryOnce.Do(func() {
		sum := sha256.Sum256(defaultSegoDictionaryGzip)
		name := "go-postery-sego-dictionary-" + hex.EncodeToString(sum[:8]) + ".txt"
		defaultDictionaryPath = filepath.Join(os.TempDir(), name)

		if _, err := os.Stat(defaultDictionaryPath); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			defaultDictionaryErr = err
			return
		}

		defaultDictionaryErr = inflateDefaultSegoDictionary(defaultDictionaryPath)
	})

	if defaultDictionaryErr != nil {
		return "", defaultDictionaryErr
	}
	return defaultDictionaryPath, nil
}

func inflateDefaultSegoDictionary(path string) error {
	reader, err := gzip.NewReader(bytes.NewReader(defaultSegoDictionaryGzip))
	if err != nil {
		return err
	}
	defer reader.Close()

	tmpFile, err := os.CreateTemp(filepath.Dir(path), ".go-postery-sego-dictionary-*")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err = io.Copy(tmpFile, reader); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err = tmpFile.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpPath, path); err != nil {
		if _, statErr := os.Stat(path); statErr == nil {
			return nil
		}
		return err
	}
	return nil
}
