package web

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	gorecslices "github.com/mishaprokop4ik/gorecs-search/pkg/slices"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type Reference string

const DefaultReference = "a"

func BaseRetryPolicy() RetryPolicyFunc {
	return func(err error, response *http.Response) bool {
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

		if response.StatusCode == 0 ||
			(response.StatusCode >= 500 && response.StatusCode != http.StatusNotImplemented) {
			return true
		}

		return false
	}
}

func DefaultContentFilterOption() FilterOption {
	return FilterOption{
		Tags: []string{
			"script",
			"style",
			"img",
			"iframe",
			"noscript",
		},
		Type: FilterExclude,
	}
}

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
)

const htmlLinkTag = "a"

type Scraper struct {
	client *Client
	sites  map[string][]string
	mutex  *sync.RWMutex
}

func NewScraper() *Scraper {
	return &Scraper{client: NewClient(BaseRetryPolicy()), mutex: &sync.RWMutex{}}
}

func (s *Scraper) Scrape(baseURL string) (map[string][]string, error) {
	result := make(map[string][]string)

	if !s.client.existPage(baseURL) {
		return map[string][]string{}, fmt.Errorf("%s, url: %s", ErrPageDoesNotExist, baseURL)
	}
	/*
		TODO: check that page by this link exist
	*/
	links, err := s.pullReferences(baseURL)
	if err != nil {
		return map[string][]string{}, fmt.Errorf("failed to pull references by %s link, err: %w",
			baseURL, err)
	}

	fmt.Println("got base links", links)

	basePageContent, err := s.pullContent(baseURL)
	if err != nil {
		return map[string][]string{}, fmt.Errorf("failed to pull content from %s url, err: %w",
			baseURL, err)
	}

	result[baseURL] = basePageContent

	pagech := make(chan Page)
	defer func() { close(pagech) }()

	errch := make(chan error)
	defer func() { close(errch) }()

	stopch := make(chan struct{})
	defer func() { close(stopch) }()

	for i, link := range links {
		go func(link string, last bool) {
			content, err := s.pullContent(link)
			if err != nil {
				errch <- err
				return
			}

			page := Page{
				URL:     link,
				Content: content,
			}
			pagech <- page

			if last {
				stopch <- struct{}{}
				return
			}
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
			fmt.Println("caught an error:", err)
		case <-stopch:
			break Loop
		}
	}

	return result, nil
}

func (s *Scraper) pullContent(url string) ([]string, error) {
	if !s.client.existPage(url) {
		return []string{}, fmt.Errorf("%s, url: %s", ErrPageDoesNotExist, url)
	}

	resp, err := s.client.Get(url)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("cannot fetch page %s - %w", url, err)
	}
	tags := s.client.FilterPageElements(resp.Body, DefaultContentFilterOption())

	result := make([]string, len(tags))
	for i, tag := range tags {
		// TODO: choose one
		//if tag.Type == Body {
		//	result[i] = tag.Body
		//}
		if strings.ReplaceAll(tag.Body, " ", "") != "" {
			result[i] = tag.Body
		}
	}

	return result, nil
}

func (s *Scraper) pullReferences(baseURL string) ([]string, error) {
	if _, err := url.Parse(baseURL); err != nil {
		return []string{}, fmt.Errorf("incorrent url param: %w", err)
	}

	resp, err := s.client.Get(baseURL)
	if err != nil {
		return []string{}, fmt.Errorf("cannot fetch page: %s, err: %s", baseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	linkTags := s.client.FilterPageElements(resp.Body, FilterOption{
		Tags: []string{htmlLinkTag},
		Type: FilterInclude,
	})
	pageLinks := make([]string, 0)
	for _, tag := range linkTags {
		if _, ok := tag.Attributes["href"]; ok {
			link, _, _ := strings.Cut(tag.Attributes["href"], "#")
			if link == "" {
				continue
			}
			if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
				if !gorecslices.Exist(link, pageLinks) {
					pageLinks = append(pageLinks, link)
				}
			} else if strings.HasPrefix(link, "//") {
				link := strings.TrimLeft(link, "/")
				if !gorecslices.Exist(link, pageLinks) {
					pageLinks = append(pageLinks, fmt.Sprintf("https://%s", link))
				}
			} else if strings.HasPrefix(link, "/") {
				testURL, _ := url.Parse(baseURL)

				link := fmt.Sprintf("%s://%s/%s", testURL.Scheme, testURL.Host, strings.TrimLeft(link, "/"))
				if !gorecslices.Exist(link, pageLinks) {
					pageLinks = append(pageLinks, link)
				}
			}
		}
	}

	return pageLinks, nil
}
