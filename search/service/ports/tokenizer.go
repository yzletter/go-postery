package ports

type Tokenizer interface {
	Cut(string) []string
}
