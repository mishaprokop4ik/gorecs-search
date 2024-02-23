package web

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/mishaprokop4ik/gorecs-search/pkg/slices"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
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

func (t *Tag) addToken(k, v string) {
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
			for slices.Exist(string(tagName), option.Tags) {
				// handle case, when value inside tag in excluded
				// TODO: find another approach
				if string(tagName) == "" {
					tokenType = token.Next()
					tagName, _ = token.TagName()

					continue
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
		case FilterInclude:
			for !slices.Exist(string(tagName), option.Tags) {
				for tokenType != html.ErrorToken && tokenType != html.EndTagToken {
					tokenType = token.Next()
					tagName, _ = token.TagName()
				}

				if tokenType == html.EndTagToken {
					tokenType = token.Next()
					tagName, _ = token.TagName()
				}
			}
		}

		if tokenType == html.ErrorToken {
			break
		}

		tag := Tag{
			Name:       string(tagName),
			Attributes: map[string]string{},
			Body:       string(token.Text()),

			Raw: token.Raw(),
		}

		for {
			k, v, a := token.TagAttr()

			tag.addToken(string(k), string(v))

			if !a {
				break
			}
		}

		tags = append(tags, tag)
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

type Scraper struct {
	client *Client
	sites  map[string][]string

	mutex *sync.RWMutex
}

func NewScraper() *Scraper {
	return &Scraper{client: NewClient(nil), mutex: &sync.RWMutex{}}
}

func (s *Scraper) Scrape(baseURL string) (map[string][]string, error) {
	result := make(map[string][]string)

	if !s.existPage(baseURL) {
		return map[string][]string{}, fmt.Errorf("%s, url: %s", ErrorPageDoesNotExist, baseURL)
	}

	links, err := s.pullReferences(baseURL)
	if err != nil {
		return map[string][]string{}, fmt.Errorf("cannot get ")
	}

	pagech := make(chan Page)
	defer func() { close(pagech) }()

	errch := make(chan error)
	defer func() { close(errch) }()

	stopch := make(chan struct{})
	defer func() { close(stopch) }()

	for i, link := range links {
		go func(link string, last bool) {
			content, err := s.GetPageContent(link)
			if err != nil {
				errch <- err
				return
			}

			if last {
				stopch <- struct{}{}
			}

			page := Page{
				URL:     link,
				Content: content,
			}
			pagech <- page
		}(link, i == len(links)-1)
	}

Loop:
	for {
		select {
		case page := <-pagech:
			s.mutex.Lock()
			result[page.URL] = page.Content
			s.mutex.Unlock()
		case err := <-errch:
			fmt.Printf("caught an error: %s\n", err)
		case <-stopch:
			break Loop
		}
	}

	return result, nil
}

func (s *Scraper) GetPageContent(url string) ([]string, error) {
	if !s.existPage(url) {
		return []string{}, fmt.Errorf("%s, url: %s", ErrorPageDoesNotExist, url)
	}

	terms, err := s.client.GetPageContent(url)
	if err != nil {
		panic(fmt.Sprintf("caught unexpected error - %s", err))
	}
	return terms, nil
}

func (s *Scraper) pullReferences(url string) ([]string, error) {

	//s.client.FilterPageElements()
	panic("")
}

func (s *Scraper) existPage(url string) bool {
	resp, err := http.Get(url)

	if errors.Is(err, context.DeadlineExceeded) {
		// make 5 calls
	}
	return !(resp.StatusCode == http.StatusNotFound)
}

func baseRetryPolicy(err error, response *http.Response) bool {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return true
		}

		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false
			}

			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if notTrustedErrorRe.MatchString(v.Error()) {
				return false
			}
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false
			}
		}
	}

	if response.StatusCode == http.StatusNotFound {
		return false
	}

	if response.StatusCode == http.StatusTooManyRequests {
		return true
	}

	if response.StatusCode == 0 ||
		(response.StatusCode >= 500 && response.StatusCode != http.StatusNotImplemented) {
		return true
	}

	return false
}
