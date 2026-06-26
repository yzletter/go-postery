package model

// ToString 拼接 Field 和 Word
func (keyword *Keyword) ToString() string {
	if len(keyword.Word) > 0 {
		return keyword.Field + "\001" + keyword.Word
	}
	return ""
}
