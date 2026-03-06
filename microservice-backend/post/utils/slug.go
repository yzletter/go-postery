package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

var nonWord = regexp.MustCompile(`[^a-z0-9\-]`)

// Slugify 将标签转为唯一标识 + 短 Hash
func Slugify(name string) string {
	name = strings.TrimSpace(name) // 去空格
	name = strings.ToLower(name)   // 英文转小写
	args := pinyin.NewArgs()       // 中文转拼音

	var sb strings.Builder
	for k, r := range []rune(name) {
		switch {
		case unicode.Is(unicode.Han, r):
			py := pinyin.Pinyin(string(r), args)
			if len(py) > 0 && len(py[0][0]) > 0 {
				if k != 0 {
					sb.WriteRune('-')
				}
				sb.WriteString(py[0][0])
			}
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			sb.WriteRune(r)
		default:
			//sb.WriteRune('-')
		}
	}

	slug := nonWord.ReplaceAllString(sb.String(), "")

	h := sha1.Sum([]byte(slug))
	hash := hex.EncodeToString(h[:])[:6]

	return slug + "-" + hash
}
