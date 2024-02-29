package web_test

import (
	"github.com/mishaprokop4ik/gorecs-search/scraper/web"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient_GetSuccess(t *testing.T) {
	// check that it redirects automatically

	c := web.NewClient(web.BaseRetryPolicy(), 5)

	go func() {
		n := 4
		helloHandler := func(w http.ResponseWriter, req *http.Request) {
			n--
			if n == 0 {
				w.WriteHeader(http.StatusOK)
			}
			w.WriteHeader(http.StatusTooManyRequests)
		}

		http.HandleFunc("/", helloHandler)

		assert.NoError(t, http.ListenAndServe(":8081", nil))
	}()
	_, err := c.Get("http://localhost:8081/")
	assert.NoError(t, err)
}

func TestClient_GetWithError(t *testing.T) {
	// check that it redirects automatically

	c := web.NewClient(web.BaseRetryPolicy(), 5)

	go func() {
		helloHandler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		http.HandleFunc("/error", helloHandler)

		assert.NoError(t, http.ListenAndServe(":8080", nil))
	}()
	_, err := c.Get("http://localhost:8080/error")
	assert.EqualError(t, err, "cannot fetch http://localhost:8080/error page: exceeded retries: last status code 500")
}
