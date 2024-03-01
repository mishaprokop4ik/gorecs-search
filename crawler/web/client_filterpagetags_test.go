package web_test

import (
	"github.com/mishaprokop4ik/gorecs-search/crawler/web"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestClient_FilterPageElements(t *testing.T) {
	c := web.NewClient(web.BaseRetryPolicy(), 5)

	type result struct {
		tags []web.Tag
	}

	type params struct {
		page         string
		filterAction web.FilterOption
	}
	defaultPage := `<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Two Links Example</title>
			</head>
			<body>
				<h1>Two Links Example</h1>
				<p>Here are two links:</p>
				<ul>
					<li><a href="https://www.example.com">Link 1</a></li>
					<li><a href="https://www.example.org">Link 2</a></li>
				</ul>
			</body>
			</html>`

	testCases := []struct {
		name   string
		input  params
		output result
	}{
		{
			name: "should return two <a> tags with body and close tags",
			input: params{
				page: defaultPage,
				filterAction: web.FilterOption{
					Tags: []string{
						"a",
					},
					Type: web.FilterInclude,
				},
			},
			output: result{
				tags: []web.Tag{
					// first tag
					{
						Name: "a",
						Attributes: map[string]string{
							"href": "https://www.example.com",
						},
						Body: "",
						Type: web.OpenTag,
						Raw:  []byte(`<a href="https://www.example.com">`),
					},
					{
						Name:       "",
						Attributes: map[string]string{},
						Body:       "Link 1",
						Type:       web.Body,
						Raw:        []byte(`Link 1`),
					},
					{
						Name:       "a",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.CloseTag,
						Raw:        []byte(`</a>`),
					},
					// second tag
					{
						Name: "a",
						Attributes: map[string]string{
							"href": "https://www.example.org",
						},
						Body: "",
						Type: web.OpenTag,
						Raw:  []byte(`<a href="https://www.example.org">`),
					},
					{
						Name:       "",
						Attributes: map[string]string{},
						Body:       "Link 2",
						Type:       web.Body,
						Raw:        []byte(`Link 2`),
					},
					{
						Name:       "a",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.CloseTag,
						Raw:        []byte(`</a>`),
					},
				},
			},
		},
		{
			name: "should return two everything inside div",
			input: params{
				page: `<div>
				<p>Open first div</p>
				<div>
					<p>Inside second div</p>
				</div>
				<p>Close second div</p>
			</div>
			</html>`,
				filterAction: web.FilterOption{
					Tags: []string{
						"div",
					},
					Type: web.FilterInclude,
				},
			},
			output: result{
				tags: []web.Tag{
					// first tag
					{
						Name:       "div",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.OpenTag,
						Raw:        []byte(`<div>`),
					},
					{
						Name:       "p",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.OpenTag,
						Raw:        []byte(`<p>`),
					},
					{
						Name:       "",
						Attributes: map[string]string{},
						Body:       "Open first div",
						Type:       web.Body,
						Raw:        []byte(`Open first div`),
					},
					{
						Name:       "p",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.CloseTag,
						Raw:        []byte(`</p>`),
					},
					{
						Name:       "div",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.OpenTag,
						Raw:        []byte(`<div>`),
					},
					{
						Name:       "p",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.OpenTag,
						Raw:        []byte(`<p>`),
					},
					{
						Name:       "",
						Attributes: map[string]string{},
						Body:       "Inside second div",
						Type:       web.Body,
						Raw:        []byte(`Inside second div`),
					},
					{
						Name:       "p",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.CloseTag,
						Raw:        []byte(`</p>`),
					},
					{
						Name:       "div",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.CloseTag,
						Raw:        []byte(`</div>`),
					},
					{
						Name:       "div",
						Attributes: map[string]string{},
						Body:       "",
						Type:       web.CloseTag,
						Raw:        []byte(`</div>`),
					},
				},
			},
		},
		{
			name: "should return DOCTYPE html after excluding html tag",
			input: params{
				page: defaultPage,
				filterAction: web.FilterOption{
					Tags: []string{
						"html",
					},
					Type: web.FilterExclude,
				},
			},
			output: result{
				tags: []web.Tag{
					{
						Name:       "DOCTYPE",
						Attributes: map[string]string{},
						Raw:        []byte("<!DOCTYPE html>"),
						Body:       "html",
						Type:       web.Doctype,
					},
				},
			},
		},
		{
			name: "should return only DOCTYPE and tags inside head",
			input: params{
				page: defaultPage,
				filterAction: web.FilterOption{
					Tags: []string{
						"body",
					},
					Type: web.FilterExclude,
				},
			},
			output: result{
				tags: []web.Tag{
					{
						Name:       "DOCTYPE",
						Attributes: map[string]string{},
						Raw:        []byte("<!DOCTYPE html>"),
						Body:       "html",
						Type:       web.Doctype,
					},
					{
						Name: `html`,
						Attributes: map[string]string{
							"lang": "en",
						},
						Raw:  []byte(`<html lang="en">`),
						Body: "",
						Type: web.OpenTag,
					},
					{
						Name:       `head`,
						Attributes: map[string]string{},
						Raw:        []byte(`<head>`),
						Body:       "",
						Type:       web.OpenTag,
					},
					{
						Name: `meta`,
						Attributes: map[string]string{
							"charset": "UTF-8",
						},
						Raw:  []byte(`<meta charset="UTF-8">`),
						Body: "",
						Type: web.OpenTag,
					},
					{
						Name: `meta`,
						Attributes: map[string]string{
							"name":    "viewport",
							"content": "width=device-width, initial-scale=1.0",
						},
						Raw:  []byte(`<meta name="viewport" content="width=device-width, initial-scale=1.0">`),
						Body: "",
						Type: web.OpenTag,
					},
					{
						Name:       `title`,
						Attributes: map[string]string{},
						Raw:        []byte(`<title>`),
						Body:       "",
						Type:       web.OpenTag,
					},
					{
						Name:       ``,
						Attributes: map[string]string{},
						Raw:        []byte(`Two Links Example`),
						Body:       "Two Links Example",
						Type:       web.Body,
					},
					{
						Name:       `title`,
						Attributes: map[string]string{},
						Raw:        []byte(`</title>`),
						Body:       "",
						Type:       web.CloseTag,
					},
					{
						Name:       `head`,
						Attributes: map[string]string{},
						Raw:        []byte(`</head>`),
						Body:       "",
						Type:       web.CloseTag,
					},
					{
						Name:       `html`,
						Attributes: map[string]string{},
						Raw:        []byte(`</html>`),
						Body:       "",
						Type:       web.CloseTag,
					},
				},
			},
		},
		{
			name: "should return only DOCTYPE and title tag",
			input: params{
				page: defaultPage,
				filterAction: web.FilterOption{
					Tags: []string{
						"body",
						"meta",
					},
					Type: web.FilterExclude,
				},
			},
			output: result{
				tags: []web.Tag{
					{
						Name:       "DOCTYPE",
						Attributes: map[string]string{},
						Raw:        []byte("<!DOCTYPE html>"),
						Body:       "html",
						Type:       web.Doctype,
					},
					{
						Name: `html`,
						Attributes: map[string]string{
							"lang": "en",
						},
						Raw:  []byte(`<html lang="en">`),
						Body: "",
						Type: web.OpenTag,
					},
					{
						Name:       `head`,
						Attributes: map[string]string{},
						Raw:        []byte(`<head>`),
						Body:       "",
						Type:       web.OpenTag,
					}, {
						Name:       `title`,
						Attributes: map[string]string{},
						Raw:        []byte(`<title>`),
						Body:       "",
						Type:       web.OpenTag,
					},
					{
						Name:       ``,
						Attributes: map[string]string{},
						Raw:        []byte(`Two Links Example`),
						Body:       "Two Links Example",
						Type:       web.Body,
					},
					{
						Name:       `title`,
						Attributes: map[string]string{},
						Raw:        []byte(`</title>`),
						Body:       "",
						Type:       web.CloseTag,
					},
					{
						Name:       `head`,
						Attributes: map[string]string{},
						Raw:        []byte(`</head>`),
						Body:       "",
						Type:       web.CloseTag,
					},
					{
						Name:       `html`,
						Attributes: map[string]string{},
						Raw:        []byte(`</html>`),
						Body:       "",
						Type:       web.CloseTag,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			// when
			tags := c.FilterPageElements(io.NopCloser(strings.NewReader(tc.input.page)), web.FilterOption{
				Tags: tc.input.filterAction.Tags,
				Type: tc.input.filterAction.Type,
			})

			// expected
			assert.Equal(t, tc.output.tags, tags)
		})
	}
}
