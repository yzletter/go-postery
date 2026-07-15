package ports

type Tokenizer interface {
	CutSearch(string) []string
	Cut(string string) []string
}
