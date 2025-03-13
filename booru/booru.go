package booru

import (
	"errors"
	"net/http"
	"net/url"
)

type Metadata struct {
	Tags       []string `json:"tags"`
	Copyright  []string `json:"copyright"`
	Characters []string `json:"characters"`
	Artists    []string `json:"artists"`
	Hash       string   `json:"hash"`
	FromHost   string   `json:"from_host"`
	URL        string   `json:"url"`
}

type Post interface {
	MediaURL() string
	Tags() []string
	Artists() []string
	Characters() []string
	Copyright() []string
	SaveMedia(directory string, client *http.Client) error
	SaveMetadata(directory string) error
	Metadata() *Metadata
	IsImage() bool
	IsVideo() bool
	Size() uint64
}

var ErrBooruNotSupported error = errors.New("this booru is not supported")

func GetPosts(booruURL url.URL, page uint, tags string, client *http.Client) ([]Post, error) {
	switch booruURL.Hostname() {
	case "danbooru.donmai.us":
		danbooruPosts, err := GetPostsDanbooru(booruURL, page, tags, client)
		if err != nil {
			return nil, err
		}

		posts := make([]Post, len(danbooruPosts))
		for i, post := range danbooruPosts {
			posts[i] = &post
		}

		return posts, nil

	case "gelbooru.com":
		gelbooruPosts, err := GetPostsGelbooru(booruURL, page, tags, client)
		if err != nil {
			return nil, err
		}

		posts := make([]Post, len(gelbooruPosts))
		for i, post := range gelbooruPosts {
			posts[i] = &post
		}

		return posts, nil

	default:
		return nil, ErrBooruNotSupported
	}
}
