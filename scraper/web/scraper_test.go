package web_test

import (
	"github.com/mishaprokop4ik/gorecs-search/scraper/web"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScraper_Scrape(t *testing.T) {
	s := web.NewScraper()

	content, err := s.Scrape("https://en.wikipedia.org/wiki/Tf%E2%80%93idf")
	assert.NoError(t, err)
	for path, _ := range content {
		t.Log("path", path)
		//t.Log("content", content)
	}
}
