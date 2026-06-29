`dictionary.txt.gz` is the compressed default dictionary copied from
`github.com/huichen/sego/data/dictionary.txt`.

The tokenizer embeds this file so `search_service` can be deployed as a single
binary without relying on the runtime working directory or a Go module cache.
