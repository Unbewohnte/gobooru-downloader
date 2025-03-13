package booru

import (
	"Unbewohnte/gobooru-downloader/util"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DanbooruPost struct {
	MediaHash           string
	ID                  int64      `json:"id"`
	CreatedAt           time.Time  `json:"created_at"`
	UploaderID          int64      `json:"uploader_id"`
	Score               int        `json:"score"`
	Source              string     `json:"source"`
	MD5                 string     `json:"md5"`
	LastCommentBumpedAt *time.Time `json:"last_comment_bumped_at"`
	Rating              string     `json:"rating"`
	ImageWidth          int        `json:"image_width"`
	ImageHeight         int        `json:"image_height"`
	TagString           string     `json:"tag_string"`
	FavCount            int        `json:"fav_count"`
	FileExt             string     `json:"file_ext"`
	LastNotedAt         *time.Time `json:"last_noted_at"`
	ParentID            *int64     `json:"parent_id"`
	HasChildren         bool       `json:"has_children"`
	ApproverID          *int64     `json:"approver_id"`
	TagCountGeneral     int        `json:"tag_count_general"`
	TagCountArtist      int        `json:"tag_count_artist"`
	TagCountCharacter   int        `json:"tag_count_character"`
	TagCountCopyright   int        `json:"tag_count_copyright"`
	FileSize            int64      `json:"file_size"`
	UpScore             int        `json:"up_score"`
	DownScore           int        `json:"down_score"`
	IsPending           bool       `json:"is_pending"`
	IsFlagged           bool       `json:"is_flagged"`
	IsDeleted           bool       `json:"is_deleted"`
	TagCount            int        `json:"tag_count"`
	UpdatedAt           time.Time  `json:"updated_at"`
	IsBanned            bool       `json:"is_banned"`
	PixivID             *int64     `json:"pixiv_id"`
	LastCommentedAt     *time.Time `json:"last_commented_at"`
	HasActiveChildren   bool       `json:"has_active_children"`
	BitFlags            int        `json:"bit_flags"`
	TagCountMeta        int        `json:"tag_count_meta"`
	HasLarge            bool       `json:"has_large"`
	HasVisibleChildren  bool       `json:"has_visible_children"`
	MediaAsset          MediaAsset `json:"media_asset"`
	TagStringGeneral    string     `json:"tag_string_general"`
	TagStringCharacter  string     `json:"tag_string_character"`
	TagStringCopyright  string     `json:"tag_string_copyright"`
	TagStringArtist     string     `json:"tag_string_artist"`
	TagStringMeta       string     `json:"tag_string_meta"`
	FileURL             string     `json:"file_url"`
	LargeFileURL        string     `json:"large_file_url"`
	PreviewFileURL      string     `json:"preview_file_url"`
}

type MediaAsset struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MD5         string    `json:"md5"`
	FileExt     string    `json:"file_ext"`
	FileSize    int64     `json:"file_size"`
	ImageWidth  int       `json:"image_width"`
	ImageHeight int       `json:"image_height"`
	Duration    float64   `json:"duration"`
	Status      string    `json:"status"`
	FileKey     string    `json:"file_key"`
	IsPublic    bool      `json:"is_public"`
	PixelHash   string    `json:"pixel_hash"`
	Variants    []Variant `json:"variants"`
}

type Variant struct {
	Type    string `json:"type"`
	URL     string `json:"url"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	FileExt string `json:"file_ext"`
}

func GetPostsDanbooru(danbooruURL url.URL, page uint, tags string, client *http.Client) ([]DanbooruPost, error) {
	query := danbooruURL.Query()
	if page == 0 {
		page = 1
	}
	query.Set("page", fmt.Sprintf("%d", page))

	if tags != "" {
		query.Set("tags", tags)
	}
	danbooruURL.RawQuery = query.Encode()
	danbooruURL.Path = "/posts.json"

	data, err := util.GetContents(client, danbooruURL.String())
	if err != nil {
		return nil, err
	}

	var posts []DanbooruPost
	err = json.Unmarshal(data, &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (post *DanbooruPost) Tags() []string {
	return strings.Fields(post.TagStringGeneral)
}

func (post *DanbooruPost) Copyright() []string {
	return strings.Fields(post.TagStringCopyright)
}

func (post *DanbooruPost) Meta() []string {
	return strings.Fields(post.TagStringMeta)
}

func (post *DanbooruPost) Artists() []string {
	return strings.Fields(post.TagStringArtist)
}

func (post *DanbooruPost) Characters() []string {
	return strings.Fields(post.TagStringCharacter)
}

func (post *DanbooruPost) FileExtension() string {
	return post.FileExt
}

func (post *DanbooruPost) MediaURL() string {
	if post.FileURL == "" {
		// Fallback to large file URL
		if post.LargeFileURL == "" {
			// Fallback to source
			return post.Source
		}

		return post.LargeFileURL
	}

	return post.FileURL
}

func (post *DanbooruPost) SaveMedia(directory string, client *http.Client) error {
	// Get file contents
	contents, err := util.GetContents(client, post.MediaURL())
	if err != nil {
		return err
	}

	// Calculate hash
	hasher := sha256.New()
	hasher.Write(contents)
	mediaHash := hex.EncodeToString(hasher.Sum(nil))
	post.MediaHash = mediaHash

	fileExt := filepath.Ext(post.MediaURL())
	if fileExt == "" {
		fileExt = ".bin"
	}

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

	return nil
}

func (post *DanbooruPost) SaveMetadata(directory string) error {
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

func (post *DanbooruPost) IsImage() bool {
	imageExtensions := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"webp": true,
		"bmp":  true,
		"tiff": true,
	}

	return post.MediaAsset.Duration == 0.0 || imageExtensions[post.FileExt]
}

func (post *DanbooruPost) IsVideo() bool {
	videoExtensions := map[string]bool{
		"mp4":  true,
		"webm": true,
		"avi":  true,
		"mov":  true,
		"mkv":  true,
		"flv":  true,
		"wmv":  true,
	}

	return post.MediaAsset.Duration > 0.0 || videoExtensions[post.FileExt]
}

func (post *DanbooruPost) Metadata() *Metadata {
	return &Metadata{
		Tags:       post.Tags(),
		Copyright:  post.Copyright(),
		Characters: post.Characters(),
		Artists:    post.Artists(),
		Hash:       post.MediaHash,
		FromHost:   "danbooru.donmai.us",
		URL:        post.MediaURL(),
	}
}
