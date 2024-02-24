package web_test

import (
	"fmt"
	"github.com/mishaprokop4ik/gorecs-search/scraper/web"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_Ge(t *testing.T) {
	c := web.NewClient(nil)
	resp, err := c.Get("https://zetcode.com/golang/net-html/")
	defer func() { _ = resp.Body.Close() }()
	assert.Nil(t, err)

	//res := c.FilterPageElements(resp.Body, web.FilterOption{
	//	Tags: []string{"a"},
	//	Type: web.FilterInclude,
	//})
	//for _, token := range res {
	//	fmt.Println(token.Name, token.Body, token.Attributes)
	//}
	result1, err := c.GetPageContent("https://zetcode.com/golang/net-html/")
	assert.Nil(t, err)
	fmt.Println(result1)

}
