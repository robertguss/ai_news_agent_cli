package scraper

type Scraper interface {
	Scrape(url string) (string, error)
}

type JinaScraper struct{}

func (j *JinaScraper) Scrape(url string) (string, error) {
	return Scrape(url)
}

func NewJinaScraper() *JinaScraper {
	return &JinaScraper{}
}
