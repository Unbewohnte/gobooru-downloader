package booru

import (
	"Unbewohnte/gobooru-downloader/logger"
	"Unbewohnte/gobooru-downloader/util"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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

type PostInfo struct {
	Tags     []Tag  `json:"tags"`
	MediaURL string `json:"media_url"`
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

var ErrVideoPost error = errors.New("it is a video post")

// Retrieves image or video URL from post document
func getMediaURLFromDoc(postDoc *goquery.Document, hostname string, imagesOnly bool) (string, error) {
	switch hostname {
	case "danbooru.donmai.us":
		// VIDEO
		videoSrc := postDoc.Find("#content .image-container video").AttrOr("src", "")
		if videoSrc != "" && imagesOnly {
			return "", ErrVideoPost
		}
		if videoSrc != "" {
			return videoSrc, nil
		}

		// IMAGE
		// Try to get the href of .image-view-original-link
		imageOriginalLink := postDoc.Find(".image-view-original-link")
		if imageOriginalLink.Length() > 0 {
			href, exists := imageOriginalLink.Attr("href")
			if exists {
				return href, nil
			}
		}

		// Fallback to srcset of <source> inside <picture>
		imageSource := postDoc.Find("picture source")
		if imageSource.Length() > 0 {
			srcset, exists := imageSource.Attr("srcset")
			if exists {
				return srcset, nil
			}
		}

		return "", errors.New("media not found")
	default:
		return "", fmt.Errorf("%s is not supported", hostname)
	}
}

// Downloads a content from the given URL and returns its content as a byte slice along with its content-type.
func GetContents(client *http.Client, contentURL string) ([]byte, string, error) {
	response, err := util.DoGETRetry(client, contentURL)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logger.Warning("DEBUG %s", contentURL)
		return nil, "", fmt.Errorf("status code %d", response.StatusCode)
	}

	// Read the content into a byte slice
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	return data, response.Header.Get("Content-Type"), nil
}

// Saves media to directory with its hash as a name, returns its hash
func SaveMedia(client *http.Client, mediaURL string, directory string) (string, error) {
	// Get the media
	data, contentType, err := GetContents(client, mediaURL)
	if err != nil {
		return "", fmt.Errorf("failed to get media: %w", err)
	}

	// Create a SHA-256 hasher
	hasher := sha256.New()

	_, err = hasher.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Calculate the SHA-256 hash of the media content
	hashBytes := hasher.Sum(nil)
	hash := hex.EncodeToString(hashBytes)

	// Determine the file extension based on the content type
	var extension string
	switch contentType {
	// Image types
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"

	// Video types
	case "video/mp4":
		extension = ".mp4"
	case "video/webm":
		extension = ".webm"
	case "video/ogg":
		extension = ".ogg"
	case "video/quicktime":
		extension = ".mov"
	case "video/x-msvideo":
		extension = ".avi"
	case "video/x-matroska":
		extension = ".mkv"
	case "video/x-flv":
		extension = ".flv"
	default:
		// Fallback: Use the URL to get the extension
		ext := filepath.Ext(mediaURL)
		if ext != "" {
			extension = ext
		} else {
			extension = ".bin"
		}
	}

	// Create the filename using the hash and extension
	filename := filepath.Join(directory, hash+extension)

	// Create the output file
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to save media: %w", err)
	}

	return hash, nil
}

// Gets post ID, retrieves post tags and media URL
func ProcessPost(client *http.Client, postURL url.URL, imagesOnly bool) (*PostInfo, error) {
	postDoc, err := getDocument(client, postURL)
	if err != nil {
		return nil, err
	}

	tags, err := getTagsFromDoc(postDoc, postURL.Hostname())
	if err != nil {
		return nil, err
	}

	mediaURL, err := getMediaURLFromDoc(postDoc, postURL.Hostname(), imagesOnly)
	if err != nil {
		return nil, err
	}

	return &PostInfo{
		Tags:     tags,
		MediaURL: mediaURL,
	}, nil
}
