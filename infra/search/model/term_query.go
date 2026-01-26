package model

import "strings"

func NewTermQuery(field, word string) *TermQuery {
	return &TermQuery{Keyword: &Keyword{Field: field, Word: word}}
}

func (q *TermQuery) And(querys ...*TermQuery) *TermQuery {
	if len(querys) == 0 {
		return q
	}

	// 构造 Must 数组
	arr := make([]*TermQuery, 0, len(querys)+1)
	if !q.Empty() {
		arr = append(arr, q)
	}

	// 遍历每个 Query
	for _, query := range querys {
		if !query.Empty() { // 排除空 Query
			arr = append(arr, query)
		}
	}

	return &TermQuery{Must: arr}
}

func (q *TermQuery) Or(querys ...*TermQuery) *TermQuery {
	if len(querys) == 0 {
		return q
	}

	// 构造 Should 数组
	arr := make([]*TermQuery, 0, len(querys)+1)
	if !q.Empty() {
		arr = append(arr, q)
	}

	// 遍历每个 Query
	for _, query := range querys {
		if !query.Empty() { // 排除空 Query
			arr = append(arr, query)
		}
	}

	return &TermQuery{Should: arr}
}

func (q *TermQuery) Empty() bool {
	return q.Keyword == nil && len(q.Must) == 0 && len(q.Should) == 0
}

func (q *TermQuery) ToString() string {
	if q.Keyword != nil {
		return q.Keyword.ToString()
	} else if len(q.Must) > 0 {
		if len(q.Must) == 1 {
			return q.Must[0].ToString()
		}

		sb := strings.Builder{}
		sb.WriteByte('(')
		for _, e := range q.Must {
			s := e.ToString()
			if len(s) > 0 {
				sb.WriteString(s)
				sb.WriteByte('&')
			}
		}
		s := sb.String()
		s = s[0:len(s)-1] + ")"
		return s
	} else if len(q.Should) > 0 {
		if len(q.Should) == 1 {
			return q.Should[0].ToString()
		}

		sb := strings.Builder{}
		sb.WriteByte('(')
		for _, e := range q.Should {
			s := e.ToString()
			if len(s) > 0 {
				sb.WriteString(s)
				sb.WriteByte('|')
			}
		}
		s := sb.String()
		s = s[0:len(s)-1] + ")"
		return s
	}
	return ""
}
