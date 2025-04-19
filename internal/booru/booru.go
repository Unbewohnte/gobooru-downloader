/*
   gobooru-downloader
   Copyright (C) 2025 Kasyanov Nikolay Alexeevich (Unbewohnte)

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
	Size       uint64   `json:"size"`
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
