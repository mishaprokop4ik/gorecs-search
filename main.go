package main

import (
	"github.com/mishaprokop4ik/gorecs-search/scraper/web"
)

func main() {
	client := web.NewClient(web.BaseRetryPolicy(), web.DefaultContentFilterOption())

	//content, err := client.Get("https://www.nytimes.com/international/")
	//if err != nil {
	//	panic(err)
	//}

	//elements := client.FilterPageElements(content.Body, web.FilterOption{
	//	Tags: []string{
	//		"body",
	//		//"a",
	//	},
	//	Type: web.FilterInclude,
	//})
	_, err := client.GetPageContent("https://www.nytimes.com/international/")
	if err != nil {
		panic(err)
	}
	//for _, e := range content {
	//	fmt.Println(e)
	//}
}
