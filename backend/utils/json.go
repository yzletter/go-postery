package utils

// ExtractJSON 从可能包含 markdown 代码块的文本中提取 JSON
func ExtractJSON(text string) string {
	// 尝试提取 ```json ... ``` 中的内容
	start := -1
	for i := 0; i < len(text)-2; i++ {
		if text[i] == '{' {
			start = i
			break
		}
	}
	if start == -1 {
		return text
	}

	// 找到最后一个 }
	end := -1
	for i := len(text) - 1; i >= start; i-- {
		if text[i] == '}' {
			end = i + 1
			break
		}
	}
	if end == -1 {
		return text
	}

	return text[start:end]
}
