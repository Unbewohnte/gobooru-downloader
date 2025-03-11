package booru

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Tag struct {
	IsArtist    bool   `json:"is_artist"`
	IsCharacter bool   `json:"is_character"`
	Value       string `json:"value"`
}

func NewTag(value string, isArtist bool, isCharacter bool) Tag {
	return Tag{
		IsArtist:    isArtist,
		IsCharacter: isCharacter,
		Value:       value,
	}
}

// Retrieves tags from booru post page document
func getTagsFromDoc(document *goquery.Document, hostname string) ([]Tag, error) {
	var tags []Tag

	switch hostname {
	case "danbooru.donmai.us":
		// Extract artist tags
		document.Find("#tag-list ul.artist-tag-list li span a.search-tag").Each(func(i int, s *goquery.Selection) {
			tags = append(tags, NewTag(s.Text(), true, false))
		})

		// Extract character tags
		document.Find("#tag-list ul.character-tag-list li span a.search-tag").Each(func(i int, s *goquery.Selection) {
			tags = append(tags, NewTag(s.Text(), false, true))
		})

		// Extract general tags
		document.Find("#tag-list ul.general-tag-list li span a.search-tag").Each(func(i int, s *goquery.Selection) {
			tags = append(tags, NewTag(s.Text(), false, false))
		})

	default:
		return nil, fmt.Errorf("%s is not supported", hostname)
	}

	return tags, nil
}

func GetTags(client *http.Client, postURL url.URL) ([]Tag, error) {
	doc, err := getDocument(client, postURL)
	if err != nil {
		return nil, err
	}

	return getTagsFromDoc(doc, postURL.Hostname())
}

func postIdFromText(text string) string {
	re := regexp.MustCompile(`ID: (\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// Returns post ID from post document. Returns error if not post ID was found
func GetPostIdFromDoc(postDoc *goquery.Document, hostname string) (string, error) {
	switch hostname {
	case "danbooru.donmai.us":
		postInfo := postDoc.Find("#post-information li:contains('ID:')").Text()
		postID := postIdFromText(postInfo)
		if postID != "" {
			return postID, nil
		} else {
			return "", errors.New("no post ID found")
		}
	default:
		return "", fmt.Errorf("%s is not supported", hostname)
	}
}

type PostInfo struct {
	PostID   string
	Tags     []Tag
	ImageURL string
}

// Retrieves image URL from post document
func getImageURLFromDoc(postDoc *goquery.Document, hostname string) (string, error) {
	switch hostname {
	case "danbooru.donmai.us":
		// Try to get the href of .image-view-original-link
		originalLink := postDoc.Find(".image-view-original-link")
		if originalLink.Length() > 0 {
			href, exists := originalLink.Attr("href")
			if exists {
				return href, nil
			}
		}

		// Fallback to srcset of <source> inside <picture>
		source := postDoc.Find("picture source")
		if source.Length() > 0 {
			srcset, exists := source.Attr("srcset")
			if exists {
				return srcset, nil
			}
		}

		return "", errors.New("image not found")
		// source, exists := postDoc.Find("picture source").Attr("srcset")
		// if !exists {
		// 	return "", errors.New("image not found")
		// }
		// return source, nil
	default:
		return "", fmt.Errorf("%s is not supported", hostname)
	}
}

func SaveImage(client *http.Client, imageURL string, filename string) error {
	response, err := client.Get(imageURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch image: status code %d", response.StatusCode)
	}

	// Get the content type from the response header
	contentType := response.Header.Get("Content-Type")
	var extension string

	// Determine the file extension based on the content type
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	default:
		// Fallback: Use the URL to get the extension
		ext := filepath.Ext(imageURL)
		if ext != "" {
			extension = ext
		} else {
			return fmt.Errorf("unsupported image type: %s", contentType)
		}
	}

	if !strings.HasSuffix(filename, extension) {
		filename += extension
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

// Gets post ID, retrieves post tags and image URL
func ProcessPost(client *http.Client, postURL url.URL) (*PostInfo, error) {
	postDoc, err := getDocument(client, postURL)
	if err != nil {
		return nil, err
	}

	postID, err := GetPostIdFromDoc(postDoc, postURL.Hostname())
	if err != nil {
		return nil, err
	}

	tags, err := getTagsFromDoc(postDoc, postURL.Hostname())
	if err != nil {
		return nil, err
	}

	imageURL, err := getImageURLFromDoc(postDoc, postURL.Hostname())
	if err != nil {
		return nil, err
	}

	return &PostInfo{
		PostID:   postID,
		Tags:     tags,
		ImageURL: imageURL,
	}, nil
}
