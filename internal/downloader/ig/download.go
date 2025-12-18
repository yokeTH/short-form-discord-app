package ig

import (
	"fmt"
	"io"
	"net/http"
)

const API_URL = "https://www.instagram.com/graphql/query"

func DownloadInstragramVideo(url string) ([]byte, error) {
	postID, ok := parseInstagramVideoURL(url)
	if !ok {
		return nil, ErrUnsupportURL
	}

	post, err := fetchInstagramPost(postID)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch post: %v", err)
	}

	if !post.Data.XDTShortcodeMedia.IsVideo || post.Data.XDTShortcodeMedia.VideoURL == "" {
		return nil, ErrNotVideo
	}

	resp, err := http.Get(post.Data.XDTShortcodeMedia.VideoURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch video: %v", err)
	}
	defer resp.Body.Close()

	videoData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read video data: %v", err)
	}

	return videoData, nil
}
