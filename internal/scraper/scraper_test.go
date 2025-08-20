package scraper

import (
        "io"
        "net"
        "net/http"
        "net/url"
        "strings"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
        handler func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
        return m.handler(req)
}

func TestScrape(t *testing.T) {
        t.Run("happy path - returns markdown content", func(t *testing.T) {
                expectedMarkdown := "# Test Article\n\nThis is test content from Jina Reader."
                
                originalClient := HTTPClient
                HTTPClient = &http.Client{
                        Transport: &mockRoundTripper{
                                handler: func(req *http.Request) (*http.Response, error) {
                                        assert.Equal(t, "GET", req.Method)
                                        assert.Equal(t, "text/markdown", req.Header.Get("Accept"))
                                        assert.Equal(t, "https://r.jina.ai/https://example.com/article", req.URL.String())
                                        
                                        return &http.Response{
                                                StatusCode: http.StatusOK,
                                                Body:       io.NopCloser(strings.NewReader(expectedMarkdown)),
                                                Header:     make(http.Header),
                                        }, nil
                                },
                        },
                }
                defer func() { HTTPClient = originalClient }()

                result, err := Scrape("https://example.com/article")
                
                require.NoError(t, err)
                assert.Equal(t, expectedMarkdown, result)
        })

        t.Run("invalid URL format", func(t *testing.T) {
                testCases := []string{
                        ":::",
                        "not-a-url",
                        "ftp://example.com",
                        "",
                        "javascript:alert('xss')",
                }

                for _, invalidURL := range testCases {
                        t.Run(invalidURL, func(t *testing.T) {
                                result, err := Scrape(invalidURL)
                                
                                assert.Empty(t, result)
                                assert.ErrorIs(t, err, ErrInvalidURL)
                        })
                }
        })

        t.Run("HTTP error status codes", func(t *testing.T) {
                testCases := []struct {
                        statusCode int
                        name       string
                }{
                        {http.StatusNotFound, "404 Not Found"},
                        {http.StatusInternalServerError, "500 Internal Server Error"},
                        {http.StatusBadRequest, "400 Bad Request"},
                        {http.StatusForbidden, "403 Forbidden"},
                }

                for _, tc := range testCases {
                        t.Run(tc.name, func(t *testing.T) {
                                originalClient := HTTPClient
                                HTTPClient = &http.Client{
                                        Transport: &mockRoundTripper{
                                                handler: func(req *http.Request) (*http.Response, error) {
                                                        return &http.Response{
                                                                StatusCode: tc.statusCode,
                                                                Body:       io.NopCloser(strings.NewReader("")),
                                                                Header:     make(http.Header),
                                                        }, nil
                                                },
                                        },
                                }
                                defer func() { HTTPClient = originalClient }()

                                result, err := Scrape("https://example.com/article")
                                
                                assert.Empty(t, result)
                                require.Error(t, err)
                                
                                var statusErr ErrStatus
                                assert.ErrorAs(t, err, &statusErr)
                                assert.Equal(t, tc.statusCode, statusErr.Code)
                        })
                }
        })

        t.Run("network error", func(t *testing.T) {
                originalClient := HTTPClient
                HTTPClient = &http.Client{
                        Transport: &mockRoundTripper{
                                handler: func(req *http.Request) (*http.Response, error) {
                                        return nil, &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Name: "nonexistent.invalid"}}
                                },
                        },
                }
                defer func() { HTTPClient = originalClient }()

                result, err := Scrape("https://nonexistent.invalid/article")
                
                assert.Empty(t, result)
                assert.Error(t, err)
                assert.NotErrorIs(t, err, ErrInvalidURL)
        })

        t.Run("request timeout", func(t *testing.T) {
                originalClient := HTTPClient
                HTTPClient = &http.Client{
                        Timeout: 100 * time.Millisecond,
                        Transport: &mockRoundTripper{
                                handler: func(req *http.Request) (*http.Response, error) {
                                        select {
                                        case <-req.Context().Done():
                                                return nil, req.Context().Err()
                                        case <-time.After(200 * time.Millisecond):
                                                return &http.Response{
                                                        StatusCode: http.StatusOK,
                                                        Body:       io.NopCloser(strings.NewReader("too slow")),
                                                        Header:     make(http.Header),
                                                }, nil
                                        }
                                },
                        },
                }
                defer func() { HTTPClient = originalClient }()

                result, err := Scrape("https://example.com/article")
                
                assert.Empty(t, result)
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "deadline exceeded")
        })

        t.Run("empty response body", func(t *testing.T) {
                originalClient := HTTPClient
                HTTPClient = &http.Client{
                        Transport: &mockRoundTripper{
                                handler: func(req *http.Request) (*http.Response, error) {
                                        return &http.Response{
                                                StatusCode: http.StatusOK,
                                                Body:       io.NopCloser(strings.NewReader("")),
                                                Header:     make(http.Header),
                                        }, nil
                                },
                        },
                }
                defer func() { HTTPClient = originalClient }()

                result, err := Scrape("https://example.com/article")
                
                require.NoError(t, err)
                assert.Empty(t, result)
        })
}

func TestBuildJinaURL(t *testing.T) {
        testCases := []struct {
                input    string
                expected string
                name     string
        }{
                {
                        input:    "https://example.com/article",
                        expected: "https://r.jina.ai/https://example.com/article",
                        name:     "HTTPS URL",
                },
                {
                        input:    "http://example.com/article",
                        expected: "https://r.jina.ai/http://example.com/article",
                        name:     "HTTP URL",
                },
                {
                        input:    "https://example.com/path/to/article?param=value",
                        expected: "https://r.jina.ai/https://example.com/path/to/article?param=value",
                        name:     "URL with path and query",
                },
        }

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        u, err := url.ParseRequestURI(tc.input)
                        require.NoError(t, err)
                        
                        result := buildJinaURL(u)
                        assert.Equal(t, tc.expected, result)
                })
        }
}
