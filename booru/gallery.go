package booru

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// Retrieves post URLs from booru gallery page
func retrievePostsFromDoc(document *goquery.Document, hostname string) ([]string, error) {
	var urls []string = []string{}

	switch hostname {
	case "danbooru.donmai.us":
		document.Find(".post-preview-link").Each(func(i int, s *goquery.Selection) {
			urls = append(urls, s.AttrOr("href", ""))
		})
	case "gelbooru.com":
		document.Find("").Each(func(i int, s *goquery.Selection) {
			urls = append(urls, s.AttrOr("href", ""))
		})
	default:
		return nil, fmt.Errorf("%s is not supported", hostname)
	}

	return urls, nil
}

// Retrieves all found posts from a gallery page
func GetPosts(client *http.Client, galleryURL url.URL) ([]string, error) {
	doc, err := getDocument(client, galleryURL)
	if err != nil {
		return nil, err
	}

	return retrievePostsFromDoc(doc, galleryURL.Hostname())
}
