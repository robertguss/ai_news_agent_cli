package scraper

type MockScraper struct {
	content string
	err     error
}

func NewMockScraper(content string, err error) *MockScraper {
	return &MockScraper{
		content: content,
		err:     err,
	}
}

func (m *MockScraper) Scrape(url string) (string, error) {
	return m.content, m.err
}
