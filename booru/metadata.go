package booru

import (
	"encoding/json"
	"os"
)

type FileMetadata struct {
	Post      PostInfo `json:"post"`
	ImageHash string   `json:"image_hash"`
	FromHost  string   `json:"from_host"`
}

func NewMetadata(post PostInfo, fromHost string, imageHash string) FileMetadata {
	return FileMetadata{
		Post:      post,
		FromHost:  fromHost,
		ImageHash: imageHash,
	}
}

func SaveMetadataJson(metadata FileMetadata, filename string) error {
	metadataFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer metadataFile.Close()

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = metadataFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}
