package web_test

import (
	"fmt"
	"github.com/mishaprokop4ik/gorecs-search/scraper/web"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScraper_PullReferences(t *testing.T) {
	s := web.NewScraper()

	links, err := s.PullReferences("https://stackoverflow.com/questions/73031647/why-does-len-on-x-net-html-token-attr-return-a-non-zero-value-for-an-empty-sli")
	assert.Nil(t, err)

	for l := range links {
		fmt.Println(links[l])
	}
}
