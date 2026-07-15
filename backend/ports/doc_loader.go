package ports

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

var (
	ErrFileNotExist   = errors.New("loader : 文件不存在")
	ErrFileNotSupport = errors.New("loader : 文件不支持")
	ErrFileReadFailed = errors.New("loader : 读取文件失败")
	ErrFileEmpty      = errors.New("loader : 文件为空")
)

type DocLoader interface {
	// LoadFile 加载 PDF / TXT / DOCX / Markdown 文件并解析
	LoadFile(ctx context.Context, path string) (string, error)
}

type MyDocLoader struct {
}

func NewMyDocLoader() DocLoader {
	return &MyDocLoader{}
}

func (loader *MyDocLoader) LoadFile(ctx context.Context, path string) (string, error) {
	// 判断文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", ErrFileNotExist
	}

	// 获取后缀拓展类型并转为小写
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".markdown":
		return loadText(path)
	case ".pdf":
		return loadPDF(path)
	case ".docx":
		return loadDOCX(path)
	case ".doc":
		return "", ErrFileNotSupport
	default:
		// 尝试当纯文本读取
		return loadText(path)
	}
}

// loadText 读取纯文本文件
func loadText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ErrFileReadFailed
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", ErrFileEmpty
	}
	return content, nil
}

func loadPDF(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", ErrFileReadFailed
	}
	defer file.Close()

	var sb strings.Builder
	totalPage := reader.NumPage()
	if totalPage == 0 {
		return "", ErrFileEmpty
	}

	for i := 1; i <= totalPage; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	content := strings.TrimSpace(sb.String())
	if content == "" {
		return "", ErrFileEmpty
	}
	return content, nil
}

// loadDOCX 解析 DOCX 文件，提取纯文本
func loadDOCX(path string) (string, error) {
	reader, err := docx.ReadDocxFile(path)
	if err != nil {
		return "", ErrFileReadFailed
	}
	defer reader.Close()

	doc := reader.Editable()
	content := doc.GetContent()

	// docx 库返回的内容可能含 XML 标签残留，做基本清理
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ErrFileEmpty
	}

	return content, nil
}
