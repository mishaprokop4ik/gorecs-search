package main

import (
	"fmt"
	"github.com/mishaprokop4ik/gorecs-search/crawler/web"
	"github.com/mishaprokop4ik/gorecs-search/lexer"
	"github.com/mishaprokop4ik/gorecs-search/ranker"
)

func main() {
	s := web.NewCrawler()

	content, err := s.Scrape("https://go.dev/learn/")
	if err != nil {
		panic(err)
	}

	r := ranker.NewModel(map[string][]string{})
	for url, content := range content {
		l := lexer.NewLexer(content...)
		contentTokens := l.All()
		r.AddDocuments(map[string][]string{
			url: contentTokens,
		})
	}

	fmt.Println(r.Rank("by", "examples"))
}
