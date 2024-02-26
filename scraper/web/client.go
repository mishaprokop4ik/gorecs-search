package web

import (
	"errors"
	"fmt"
	gorecslices "github.com/mishaprokop4ik/gorecs-search/pkg/slices"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

var ErrPageDoesNotExist = errors.New("page doesn't exist")

type Page struct {
	Type    string
	URL     string
	Content []string
}

type Client struct {
	httpClient http.Client

	retryPolicy func(err error, resp *http.Response) bool
}

type TagType int

const (
	Doctype TagType = iota
	SelfCloseTag
	OpenTag
	Body
	CloseTag
)

// String returns a string representation of the TokenType.
func (t TagType) String() string {
	switch t {
	case Doctype:
		return "Doctype"
	case Body:
		return "Body"
	case OpenTag:
		return "OpenTag"
	case CloseTag:
		return "CloseTag"
	case SelfCloseTag:
		return "SelfCloseTag"
	}

	return "Invalid(" + strconv.Itoa(int(t)) + ")"
}

type Tag struct {
	Name       string
	Body       string
	Type       TagType
	Raw        []byte
	Attributes map[string]string
}

func (t *Tag) provideType(tokenType html.TokenType) {
	switch tokenType {
	case html.DoctypeToken:
		t.Type = Doctype
		t.Name = "DOCTYPE"
	case html.SelfClosingTagToken:
		t.Type = SelfCloseTag
	case html.StartTagToken:
		t.Type = OpenTag
	case html.TextToken:
		t.Type = Body
	case html.EndTagToken:
		t.Type = CloseTag
	}
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
	for tokenType := token.Next(); tokenType != html.ErrorToken; tokenType = token.Next() {
		tagName, _ := token.TagName()
		switch option.Type {
		case FilterExclude:
			if gorecslices.Exist(string(tagName), option.Tags) {
				currentTag := string(tagName)
				for {
					tokenType = token.Next()
					tagName, _ = token.TagName()
					if tokenType == html.ErrorToken {
						break
					}

					if tokenType == html.EndTagToken && currentTag == string(tagName) {
						tokenType = token.Next()
						tagName, _ = token.TagName()
						break
					} else if (tokenType == html.SelfClosingTagToken || tokenType == html.DoctypeToken || string(tagName) == "meta") &&
						currentTag == string(tagName) {
						// TODO: check all tags and find them that have the same signature as meta
						// TODO: find possible another approach for meta tag
						tokenType = token.Next()
						tagName, _ = token.TagName()
						break
					}
				}
			}
			rawToken := slices.Clone(token.Raw())
			if strings.TrimSpace(string(rawToken)) == "" {
				continue
			}
			tag := Tag{
				Name:       string(tagName),
				Attributes: map[string]string{},
				Body:       string(token.Text()),
				Raw:        rawToken,
			}

			tag.provideType(tokenType)

			for {
				k, v, a := token.TagAttr()
				tag.addAttribute(string(k), string(v))

				if !a {
					break
				}
			}

			tags = append(tags, tag)
		case FilterInclude:
			if gorecslices.Exist(string(tagName), option.Tags) {
				for tokenType != html.EndTagToken {
					tag := Tag{
						Name:       string(tagName),
						Attributes: map[string]string{},
						Body:       string(token.Text()),
						Raw:        slices.Clone(token.Raw()),
					}

					tag.provideType(tokenType)

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
					Raw:        slices.Clone(token.Raw()),
				}

				tag.provideType(tokenType)

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
