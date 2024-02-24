package web

import (
	"errors"
	"fmt"
	"github.com/mishaprokop4ik/gorecs-search/pkg/slices"
	"golang.org/x/net/html"
	"io"
	"net/http"
	slices2 "slices"
	"strings"
	"time"
)

var ErrorPageDoesNotExist = errors.New("page doesn't exist")

type Page struct {
	Type    string
	URL     string
	Content []string
}

type Client struct {
	httpClient http.Client

	retryPolicy func(err error, resp *http.Response) bool
}

func NewTag(t string) *Tag {
	tag := &Tag{}

	return tag
}

type Tag struct {
	Name       string
	Attributes map[string]string
	Body       string

	Raw []byte
}

func (t *Tag) addAttribute(k, v string) {
	if k == "" {
		return
	}

	if t.Attributes == nil {
		t.Attributes = map[string]string{}
	}

	t.Attributes[k] = v
}

type FilterType int

const (
	FilterInclude FilterType = iota
	FilterExclude
)

type FilterOption struct {
	Tags []string
	Type FilterType
}

func NewClient(retryPolicy func(err error, resp *http.Response) bool) *Client {
	if retryPolicy == nil {
		retryPolicy = baseRetryPolicy
	}

	httpClient := http.Client{
		Timeout: 3 * time.Second,
	}
	return &Client{httpClient: httpClient, retryPolicy: retryPolicy}
}

func (c *Client) FilterPageElements(body io.ReadCloser, option FilterOption) []Tag {
	token := html.NewTokenizer(body)
	tags := make([]Tag, 0)
Root:
	for tokenType := token.Next(); tokenType != html.ErrorToken; tokenType = token.Next() {
		tagName, _ := token.TagName()
		switch option.Type {
		case FilterExclude:
			// loop for handling sequence of  tags: e.g. script, script, script
			for slices.Exist(string(tagName), option.Tags) {
				// handle case, when value inside tag in excluded
				// TODO: find another approach
				if string(tagName) == "" {
					tokenType = token.Next()
					tagName, _ = token.TagName()

					continue Root
				}

				for tokenType != html.ErrorToken && tokenType != html.EndTagToken {
					tokenType = token.Next()
					tagName, _ = token.TagName()
				}

				if tokenType == html.EndTagToken {
					tokenType = token.Next()
					tagName, _ = token.TagName()
				}
			}
			tag := Tag{
				Name:       string(tagName),
				Attributes: map[string]string{},
				Body:       string(token.Text()),

				Raw: slices2.Clone(token.Raw()),
			}

			for {
				k, v, a := token.TagAttr()

				tag.addAttribute(string(k), string(v))

				if !a {
					break
				}
			}

			tags = append(tags, tag)
		case FilterInclude:
			if slices.Exist(string(tagName), option.Tags) {
				for tokenType != html.EndTagToken {
					tag := Tag{
						Name:       string(tagName),
						Attributes: map[string]string{},
						Body:       string(token.Text()),

						Raw: slices2.Clone(token.Raw()),
					}

					for {
						k, v, a := token.TagAttr()

						tag.addAttribute(string(k), string(v))

						if !a {
							break
						}
					}
					tags = append(tags, tag)
					tokenType = token.Next()
				}

				tag := Tag{
					Name:       string(tagName),
					Attributes: map[string]string{},
					Body:       string(token.Text()),

					Raw: slices2.Clone(token.Raw()),
				}

				for {
					k, v, a := token.TagAttr()

					tag.addAttribute(string(k), string(v))

					if !a {
						break
					}
				}
				tags = append(tags, tag)
			}
		}

		if tokenType == html.ErrorToken {
			break
		}
	}

	return tags
}

func (c *Client) GetPageContent(url string) ([]string, error) {
	resp, err := c.Get(url)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, fmt.Errorf("cannot fetch page %s - %w", url, err)
	}
	tags := c.FilterPageElements(resp.Body, FilterOption{
		Tags: []string{"script", "style"},
		Type: FilterExclude,
	})

	result := make([]string, len(tags))
	for i, tag := range tags {
		if strings.ReplaceAll(tag.Body, " ", "") != "" {
			result[i] = tag.Body
		}
	}

	return result, nil
}

func (c *Client) Get(url string) (*http.Response, error) {
	resp, err := c.httpClient.Get(url)
	// TODO: implement retry mechanism
	if err != nil {
		return nil, fmt.Errorf("cannot fetch page %s - %w", url, err)
	}

	return resp, nil
}
