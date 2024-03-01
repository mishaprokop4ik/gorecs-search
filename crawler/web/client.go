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

var (
	ErrPageDoesNotExist = errors.New("page doesn't exist")
	ErrRetriesExceeded  = errors.New("exceeded retries")
)

type Page struct {
	Type    string
	URL     string
	Content []string
}

// Client provides API to collect Web data.
type Client struct {
	httpClient    http.Client
	contentFilter FilterOption

	retryPolicy   func(err error, resp *http.Response) bool
	retryAttempts uint
}

// TagType represents tag type
type TagType int

const (
	// Doctype is <!DOCTYPE ...> kind of Tag.
	Doctype TagType = iota
	// SelfCloseTag is <br /> kind of Tag.
	SelfCloseTag
	// OpenTag is <a> kind of Tag.
	OpenTag
	// Body is text inside an HTML Tag.
	Body
	// CloseTag is <a/> kind of Tag.
	CloseTag
)

// String returns a string representation of the TagType.
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

// TODO: maybe Tag type should be refactored to
// type Tag struct {
//	Open, Close       string
//	Body string
//  Attributes map[string]string
//  InnerTags []Tag
//}

// Tag contains information about Tag(usually HTML one).
type Tag struct {
	// Name is a value inside <> braces.
	Name string
	// Data inside Tag. Is always empty when Type is either: OpenTag, CloseTag, SelfCloseTag.
	Body       string
	Type       TagType
	Raw        []byte
	Attributes map[string]string
}

// provideType converts html.TokenType to internal TagType.
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
	_ FilterType = iota
	FilterInclude
	FilterExclude
)

type FilterOption struct {
	Tags []string
	Type FilterType
}

func (f *FilterOption) Empty() bool {
	return f.Tags == nil && f.Type == 0
}

type RetryPolicyFunc func(err error, resp *http.Response) bool

func NewClient(retryPolicy RetryPolicyFunc, retryAttempts uint) *Client {
	if retryPolicy == nil {
		retryPolicy = BaseRetryPolicy()
	}

	httpClient := http.Client{
		Timeout: 3 * time.Second,
	}
	return &Client{httpClient: httpClient, retryPolicy: retryPolicy, retryAttempts: retryAttempts}
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
					if tokenType == html.ErrorToken {
						break
					}
					// TODO: check all tags and find them that have the same signature as meta
					// TODO: find possible another approach for meta tag
					if gorecslices.Exist(string(tagName), option.Tags) && tokenType == html.StartTagToken && (string(tagName) == "meta" || string(tagName) == "img") {
						tokenType = token.Next()
						tagName, _ = token.TagName()
						if !gorecslices.Exist(string(tagName), option.Tags) {
							break
						}
					}
					if tokenType == html.EndTagToken && currentTag == string(tagName) {
						tokenType = token.Next()
						tagName, _ = token.TagName()
						if !gorecslices.Exist(string(tagName), option.Tags) {
							break
						}
					} else if (tokenType == html.SelfClosingTagToken || tokenType == html.DoctypeToken) &&
						currentTag == string(tagName) {
						tokenType = token.Next()
						tagName, _ = token.TagName()
						if !gorecslices.Exist(string(tagName), option.Tags) {
							break
						}
					}

					tokenType = token.Next()
					tagName, _ = token.TagName()
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
				tagCount := 1
				currentTag := string(tagName)
				closeTag := false
				for {
					if tokenType == html.ErrorToken {
						break
					}

					if (tokenType == html.EndTagToken || tokenType == html.SelfClosingTagToken) && currentTag == string(tagName) {
						tagCount--
					}

					if tagCount == 0 {
						closeTag = true
					}

					rawToken := slices.Clone(token.Raw())
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

					if tag.Type == Body && strings.TrimSpace(tag.Body) == "" {
						tokenType = token.Next()
						tagName, _ = token.TagName()
						continue
					}

					tags = append(tags, tag)

					if closeTag {
						break
					}

					tokenType = token.Next()
					tagName, _ = token.TagName()
					if currentTag == string(tagName) && tokenType == html.StartTagToken {
						tagCount++
					}
				}
			}
		}
	}

	return tags
}

func (c *Client) Get(url string) (*http.Response, error) {
	resp, err := c.httpClient.Get(url)
	for leftRetries := c.retryAttempts; c.retryPolicy(err, resp) && leftRetries > 0; leftRetries-- {
		resp, err = c.httpClient.Get(url)
		if leftRetries == 1 {
			if err != nil {
				err = fmt.Errorf("%w: %w", ErrRetriesExceeded, err)
			}
			if resp.StatusCode >= http.StatusBadRequest {
				if err == nil {
					err = ErrRetriesExceeded
				}
				err = fmt.Errorf("%w: last status code %d", err, resp.StatusCode)
			}
		}
	}
	// TODO: check for redirects and make it

	if err != nil {
		return nil, fmt.Errorf("cannot fetch %s page: %w", url, err)
	}

	return resp, nil
}

func (c *Client) ExistPage(url string) bool {
	resp, err := c.Get(url)

	if err != nil {
		fmt.Println(err)
		return false
	}

	return !(resp.StatusCode == http.StatusNotFound)
}
