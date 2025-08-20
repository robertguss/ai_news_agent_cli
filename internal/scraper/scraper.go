package scraper

import (
        "context"
        "errors"
        "fmt"
        "io"
        "net"
        "net/http"
        "net/url"
        "strconv"
        "time"

        "github.com/robertguss/ai-news-agent-cli/internal/config"
        "github.com/robertguss/ai-news-agent-cli/pkg/errs"
        "github.com/robertguss/ai-news-agent-cli/pkg/logging"
        "github.com/robertguss/ai-news-agent-cli/pkg/retry"
)

var (
        ErrInvalidURL = errors.New("scraper: invalid url")
)

type ErrStatus struct {
        Code int
}

func (e ErrStatus) Error() string {
        return "scraper: http status " + strconv.Itoa(e.Code)
}

var HTTPClient = &http.Client{
        Timeout: 0,
        Transport: &http.Transport{
                DialContext: (&net.Dialer{
                        Timeout: 30 * time.Second,
                }).DialContext,
                ResponseHeaderTimeout: 60 * time.Second,
        },
}

func Scrape(rawURL string) (string, error) {
        u, err := url.ParseRequestURI(rawURL)
        if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
                return "", ErrInvalidURL
        }

        jinaURL := buildJinaURL(u)
        
        ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
        defer cancel()

        req, err := http.NewRequestWithContext(ctx, http.MethodGet, jinaURL, nil)
        if err != nil {
                return "", err
        }
        req.Header.Set("Accept", "text/markdown")

        resp, err := HTTPClient.Do(req)
        if err != nil {
                return "", err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return "", ErrStatus{Code: resp.StatusCode}
        }

        body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
        if err != nil {
                return "", err
        }

        return string(body), nil
}

func ScrapeWithRetry(ctx context.Context, rawURL string, cfg *config.Config) (string, error) {
        ctx, cancel := context.WithTimeout(ctx, cfg.NetworkTimeout)
        defer cancel()

        var content string
        err := retry.DoWithCallback(ctx, cfg.RetryConfig(), func() error {
                var e error
                content, e = scrapeInternal(ctx, rawURL)
                return e
        }, func(attempt int, err error) {
                logging.Retry("scrape_url", attempt, err)
        })

        if err != nil {
                wrappedErr := errs.Wrap("scrape url "+rawURL, err)
                logging.Error("scrape_url", wrappedErr)
                return "", wrappedErr
        }

        return content, nil
}

func scrapeInternal(ctx context.Context, rawURL string) (string, error) {
        u, err := url.ParseRequestURI(rawURL)
        if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
                return "", ErrInvalidURL
        }

        jinaURL := buildJinaURL(u)
        
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, jinaURL, nil)
        if err != nil {
                return "", err
        }
        req.Header.Set("Accept", "text/markdown")

        resp, err := HTTPClient.Do(req)
        if err != nil {
                return "", err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return "", ErrStatus{Code: resp.StatusCode}
        }

        body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
        if err != nil {
                return "", err
        }

        return string(body), nil
}

func buildJinaURL(u *url.URL) string {
        return fmt.Sprintf("https://r.jina.ai/%s", u.String())
}
