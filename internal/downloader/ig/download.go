package ig

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

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
	lowerURL, found, err := findLowerQualityVideo(manifest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse DASH manifest")
		return nil, fmt.Errorf("Failed to parse DASH manifest: %v", err)
	}
	if !found {
		log.Error().Msg("No lower quality video under 8MB found in DASH manifest")
		return nil, fmt.Errorf("No video under 8MB available")
	}

	log.Info().Str("videoURL", lowerURL).Msg("Fetching lower quality video under 8MB")
	resp, err := http.Get(lowerURL)
	if err != nil {
		log.Error().Err(err).Str("videoURL", lowerURL).Msg("Failed to fetch lower quality video")
		return nil, fmt.Errorf("Failed to fetch lower quality video: %v", err)
	}
	return resp.Body, nil
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

func findLowerQualityVideo(manifest string) (string, bool, error) {
	var mpd MPD
	decoder := xml.NewDecoder(strings.NewReader(manifest))
	if err := decoder.Decode(&mpd); err != nil {
		return "", false, err
	}
	var videoReps []Representation
	for _, aset := range mpd.Period.AdaptationSets {
		for _, rep := range aset.Representations {
			if strings.HasPrefix(rep.BaseURL, "http") && (rep.MimeType == "" || strings.HasPrefix(rep.MimeType, "video")) {
				videoReps = append(videoReps, rep)
			}
		}
	}
	sort.Slice(videoReps, func(i, j int) bool {
		return videoReps[i].Bandwidth < videoReps[j].Bandwidth
	})
	for _, rep := range videoReps {
		ok, _, err := checkURLSize(rep.BaseURL)
		if err != nil {
			log.Warn().Err(err).Str("url", rep.BaseURL).Msg("Failed to check size for representation")
			continue
		}
		if ok {
			return rep.BaseURL, true, nil
		}
	}
	return "", false, nil
}
