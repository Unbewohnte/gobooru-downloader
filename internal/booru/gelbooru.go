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
	"Unbewohnte/gobooru-downloader/internal/proxy"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Attributes struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type GelbooruPost struct {
	MediaHash     string
	FileSize      uint64
	ID            int    `json:"id"`
	CreatedAt     string `json:"created_at"`
	Score         int    `json:"score"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	MD5           string `json:"md5"`
	Directory     string `json:"directory"`
	Image         string `json:"image"`
	Rating        string `json:"rating"`
	Source        string `json:"source"`
	Change        int64  `json:"change"`
	Owner         string `json:"owner"`
	CreatorID     int    `json:"creator_id"`
	ParentID      int    `json:"parent_id"`
	Sample        int    `json:"sample"`
	PreviewHeight int    `json:"preview_height"`
	PreviewWidth  int    `json:"preview_width"`
	PostTags      string `json:"tags"`
	Title         string `json:"title"`
	HasNotes      string `json:"has_notes"`
	HasComments   string `json:"has_comments"`
	FileURL       string `json:"file_url"`
	PreviewURL    string `json:"preview_url"`
	SampleURL     string `json:"sample_url"`
	SampleHeight  int    `json:"sample_height"`
	SampleWidth   int    `json:"sample_width"`
	Status        string `json:"status"`
	PostLocked    int    `json:"post_locked"`
	HasChildren   string `json:"has_children"`
}

type GelbooruJSONData struct {
	Attributes Attributes     `json:"@attributes"`
	Posts      []GelbooruPost `json:"post"`
}

func GetPostsGelbooru(gelbooruURL url.URL, page uint, tags string, client *http.Client) ([]GelbooruPost, error) {
	query := gelbooruURL.Query()
	query.Set("page", "dapi")
	query.Set("s", "post")
	query.Set("q", "index")
	query.Set("json", "1")

	if page == 0 {
		page = 1
	}
	query.Set("pid", fmt.Sprintf("%d", page))

	if tags != "" {
		query.Set("tags", tags)
	}
	gelbooruURL.RawQuery = query.Encode()
	gelbooruURL.Path = "/index.php"

	data, err := proxy.GetContents(client, gelbooruURL.String())
	if err != nil {
		return nil, err
	}

	var galleryData GelbooruJSONData
	err = json.Unmarshal(data, &galleryData)
	if err != nil {
		return nil, err
	}

	return galleryData.Posts, nil
}

func (post *GelbooruPost) Tags() []string {
	return strings.Fields(post.PostTags)
}

func (post *GelbooruPost) Copyright() []string {
	return nil
}

func (post *GelbooruPost) Meta() []string {
	return nil
}

func (post *GelbooruPost) Artists() []string {
	return nil
}

func (post *GelbooruPost) Characters() []string {
	return nil
}

func (post *GelbooruPost) FileExtension() string {
	return filepath.Ext(post.FileURL)
}

func (post *GelbooruPost) MediaURL() string {
	return post.FileURL
}

func (post *GelbooruPost) SaveMedia(directory string, client *http.Client) error {
	// Get file contents
	contents, err := proxy.GetContents(client, post.MediaURL())
	if err != nil {
		return err
	}

	// Calculate hash
	hasher := sha256.New()
	hasher.Write(contents)
	mediaHash := hex.EncodeToString(hasher.Sum(nil))
	post.MediaHash = mediaHash

	fileExt := filepath.Ext(post.MediaURL())

	// Save media
	path := filepath.Join(directory, mediaHash+fileExt)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(contents)
	if err != nil {
		return err
	}

	// Remember file size
	post.FileSize = uint64(len(contents))

	return nil
}

func (post *GelbooruPost) SaveMetadata(directory string) error {
	file, err := os.Create(
		filepath.Join(
			directory,
			fmt.Sprintf("%s_metadata.json", post.MediaHash),
		),
	)
	if err != nil {
		return err
	}
	defer file.Close()

	contents, err := json.Marshal(post.Metadata())
	if err != nil {
		return err
	}

	_, err = file.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

func (post *GelbooruPost) IsImage() bool {
	imageExtensions := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"webp": true,
		"bmp":  true,
		"tiff": true,
	}

	return imageExtensions[post.FileExtension()[1:]]
}

func (post *GelbooruPost) IsVideo() bool {
	videoExtensions := map[string]bool{
		"mp4":  true,
		"webm": true,
		"avi":  true,
		"mov":  true,
		"mkv":  true,
		"flv":  true,
		"wmv":  true,
	}

	return videoExtensions[post.FileExtension()[1:]]
}

func (post *GelbooruPost) Metadata() *Metadata {
	return &Metadata{
		Tags:       post.Tags(),
		Copyright:  post.Copyright(),
		Characters: post.Characters(),
		Artists:    post.Artists(),
		Hash:       post.MediaHash,
		FromHost:   "gelbooru.com",
		URL:        post.MediaURL(),
		Size:       post.Size(),
	}
}

func (post *GelbooruPost) Size() uint64 {
	return post.FileSize
}
