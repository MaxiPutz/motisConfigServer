package scrapper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Transitous struct {
	Name string
	Url  string
}

func GetProcesGTFSLinks() ([]Transitous, error) {
	url := "https://api.transitous.org/gtfs/"

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("Parse erro HTML: %v", err)
	}

	result := []Transitous{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, isThere := s.Attr("href")
		if isThere {
			isThere = strings.Contains(href, "gtfs")
			if isThere {
				foundElement := Transitous{
					Name: href,
					Url:  url + href,
				}
				result = append(result, foundElement)
			}
		}
	})

	return result, nil
}
