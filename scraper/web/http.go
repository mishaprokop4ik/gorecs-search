package web

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/mishaprokop4ik/gorecs-search/pkg/slices"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
)

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

	links, err := s.PullReferences(baseURL)
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
				return
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

func (s *Scraper) PullReferences(baseURL string) ([]string, error) {
	htmlLinkTag := "a"
	resp, err := s.client.Get(baseURL)
	if err != nil {
		return []string{}, fmt.Errorf("cannot fetch page: %s, err: %s", baseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	linkTags := s.client.FilterPageElements(resp.Body, FilterOption{
		Tags: []string{htmlLinkTag},
		Type: FilterInclude,
	})
	pageLinks := make([]string, len(linkTags))
	for i, tag := range linkTags {
		if _, ok := tag.Attributes["href"]; ok {
			link, _, _ := strings.Cut(tag.Attributes["href"], "#")
			/*
				TODO: check that page by this link exist
			*/
			if link == "" {
				continue
			}
			if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
				if !slices.Exist(link, pageLinks) {
					pageLinks[i] = link
				}
			} else if strings.HasPrefix(link, "/") {
				testURL, _ := url.Parse(baseURL)
				if testURL != nil {
					//fmt.Println(testURL.Scheme, testURL.Host)
				}

				link := fmt.Sprintf("%s://%s/%s", testURL.Scheme, testURL.Host, strings.TrimLeft(link, "/"))
				if !slices.Exist(link, pageLinks) {
					pageLinks[i] = link
				}
			}

			//pageLinks[i] = link
		}
	}

	return pageLinks, nil
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
