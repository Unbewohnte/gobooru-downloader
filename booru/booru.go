package booru

import (
	"Unbewohnte/gobooru-downloader/util"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// Performs GET on given url, returns goquery.Document
func getDocument(client *http.Client, url url.URL) (*goquery.Document, error) {
	response, err := util.DoGETRetry(client, url.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return goquery.NewDocumentFromReader(response.Body)
}
