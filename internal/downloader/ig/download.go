package ig

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/rs/zerolog/log"
)

const API_URL = "https://www.instagram.com/graphql/query"

const maxDiscordFileSize = 8 * 1024 * 1024

type MPD struct {
	XMLName xml.Name `xml:"MPD"`
	Period  Period   `xml:"Period"`
}

type Period struct {
	AdaptationSets []AdaptationSet `xml:"AdaptationSet"`
}

type AdaptationSet struct {
	ContentType     string           `xml:"contentType,attr"`
	Representations []Representation `xml:"Representation"`
}

type Representation struct {
	ID        string `xml:"id,attr"`
	Bandwidth int    `xml:"bandwidth,attr"`
	BaseURL   string `xml:"BaseURL"`
	MimeType  string `xml:"mimeType,attr"`
}

func DownloadInstragramVideo(url string) (io.Reader, error) {
	log.Info().Str("url", url).Msg("DownloadInstragramVideo called")
	postID, ok := parseInstagramVideoURL(url)
	if !ok {
		log.Error().Str("url", url).Msg("Unsupported Instagram URL format")
		return nil, ErrUnsupportURL
	}

	post, err := fetchInstagramPost(postID)
	if err != nil {
		log.Error().Err(err).Str("postID", postID).Msg("Failed to fetch Instagram post")
		return nil, fmt.Errorf("Failed to fetch post: %v", err)
	}

	if !post.Data.XDTShortcodeMedia.IsVideo || post.Data.XDTShortcodeMedia.VideoURL == "" {
		log.Info().Interface("post_response", post).Msg("Instagram post response")
		log.Error().Str("postID", postID).Msg("Post is not a video or missing video URL")
		return nil, ErrNotVideo
	}

	videoURL := post.Data.XDTShortcodeMedia.VideoURL
	sizeOK, size, err := checkURLSize(videoURL)
	if err != nil {
		log.Error().Err(err).Str("videoURL", videoURL).Msg("Failed to check video size")
		return nil, fmt.Errorf("Failed to check video size: %v", err)
	}
	if sizeOK {
		log.Info().Str("videoURL", videoURL).Int64("size", size).Msg("Video is within Discord size limit")
		resp, err := http.Get(videoURL)
		if err != nil {
			log.Error().Err(err).Str("videoURL", videoURL).Msg("Failed to fetch video")
			return nil, fmt.Errorf("Failed to fetch video: %v", err)
		}
		return resp.Body, nil
	}

	manifest := post.Data.XDTShortcodeMedia.DashInfo.VideoDashManifest
	if manifest == "" {
		log.Error().Str("videoURL", videoURL).Int64("size", size).Msg("Video too large and no DASH manifest available")
		return nil, fmt.Errorf("Video is larger than 8MB and no lower quality available")
	}

	log.Info().Msg("Parsing DASH manifest for lower quality video")
	mergedReader, err := processDASHAndMerge(manifest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to merge DASH streams")
		return nil, fmt.Errorf("Failed to process video: %v", err)
	}

	return mergedReader, nil
}

func checkURLSize(url string) (bool, int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	cl := resp.Header.Get("Content-Length")
	if cl == "" {
		return false, 0, fmt.Errorf("Content-Length header missing")
	}
	size, err := strconv.ParseInt(cl, 10, 64)
	if err != nil {
		return false, 0, err
	}
	return size <= maxDiscordFileSize, size, nil
}

func processDASHAndMerge(manifestContent string) (io.Reader, error) {
	var mpd MPD
	if err := xml.Unmarshal([]byte(manifestContent), &mpd); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	var audioURL string
	var videoReps []Representation

	for _, aset := range mpd.Period.AdaptationSets {
		switch aset.ContentType {
		case "audio":
			if len(aset.Representations) > 0 {
				audioURL = aset.Representations[0].BaseURL
			}
		case "video":
			videoReps = append(videoReps, aset.Representations...)
		}
	}

	if audioURL == "" {
		return nil, fmt.Errorf("no audio track found in manifest")
	}
	if len(videoReps) == 0 {
		return nil, fmt.Errorf("no video tracks found in manifest")
	}

	sort.Slice(videoReps, func(i, j int) bool {
		return videoReps[i].Bandwidth < videoReps[j].Bandwidth
	})

	targetVideoURL := videoReps[0].BaseURL
	log.Info().Int("bandwidth", videoReps[0].Bandwidth).Msg("Selected video quality")

	tempDir := os.TempDir()
	videoPath := filepath.Join(tempDir, "temp_video.mp4")
	audioPath := filepath.Join(tempDir, "temp_audio.mp4")
	outputPath := filepath.Join(tempDir, "output_merged.mp4")

	defer os.Remove(videoPath)
	defer os.Remove(audioPath)

	if err := downloadFile(targetVideoURL, videoPath); err != nil {
		return nil, fmt.Errorf("failed to download video stream: %w", err)
	}
	if err := downloadFile(audioURL, audioPath); err != nil {
		return nil, fmt.Errorf("failed to download audio stream: %w", err)
	}

	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c", "copy",
		"-map", "0:v:0",
		"-map", "1:a:0",
		"-movflags", "+faststart",
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Str("ffmpeg_output", string(output)).Msg("FFmpeg merge failed")
		return nil, fmt.Errorf("ffmpeg failed: %w", err)
	}

	resultData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}

	os.Remove(outputPath)

	return bytes.NewReader(resultData), nil
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
