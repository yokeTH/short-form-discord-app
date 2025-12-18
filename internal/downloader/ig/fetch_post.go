package ig

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type response struct {
	Data struct {
		XDTShortcodeMedia xdtShortcodeMedia `json:"xdt_shortcode_media"`
	} `json:"data"`
}

type xdtShortcodeMedia struct {
	ID          string `json:"id"`
	Shortcode   string `json:"shortcode"`
	Typename    string `json:"__typename"`
	ProductType string `json:"product_type"`

	IsVideo       bool    `json:"is_video"`
	VideoURL      string  `json:"video_url"`
	VideoDuration float64 `json:"video_duration"`
	HasAudio      bool    `json:"has_audio"`

	Owner Owner `json:"owner"`

	DashInfo DashInfo `json:"dash_info,omitzero"`

	EdgeSidecarToChildren EdgeSidecar `json:"edge_sidecar_to_children,omitzero"`
}

type Owner struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	IsPrivate bool   `json:"is_private"`
}

type DashInfo struct {
	IsDashEligible    bool   `json:"is_dash_eligible"`
	VideoDashManifest string `json:"video_dash_manifest"`
	NumberOfQualities int    `json:"number_of_qualities"`
}

type EdgeSidecar struct {
	Edges []SidecarEdge `json:"edges"`
}

type SidecarEdge struct {
	Node SidecarNode `json:"node"`
}

type SidecarNode struct {
	ID       string `json:"id"`
	Typename string `json:"__typename"`

	IsVideo  bool   `json:"is_video"`
	VideoURL string `json:"video_url"`
	HasAudio bool   `json:"has_audio"`

	DashInfo DashInfo `json:"dash_info,omitzero"`
}

func fetchInstagramPost(shortcode string) (*response, error) {
	log.Info().Str("shortcode", shortcode).Msg("Fetching Instagram post")
	body := buildPostBody(shortcode)

	req, err := http.NewRequest(
		"POST",
		"https://www.instagram.com/graphql/query",
		strings.NewReader(body),
	)
	if err != nil {
		log.Error().Err(err).Str("shortcode", shortcode).Msg("Failed to create HTTP request for Instagram post")
		return nil, err
	}

	// Required headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 11; SAMSUNG SM-G973U)")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRFToken", "IqQcANe4LUgoa0OaiXW2tFqltnjiIphK")
	req.Header.Set("X-IG-App-ID", "936619743392459")
	req.Header.Set("X-FB-Friendly-Name", "PolarisPostActionLoadPostQueryQuery")
	req.Header.Set("X-ASBD-ID", "359341")
	req.Header.Set("Referer", "https://www.instagram.com/p/"+shortcode+"/")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("shortcode", shortcode).Msg("Failed to perform HTTP request for Instagram post")
		return nil, err
	}
	defer resp.Body.Close()

	var res response
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		log.Error().Err(err).Str("shortcode", shortcode).Msg("Failed to decode Instagram post response")
		return nil, err
	}
	log.Info().Str("shortcode", shortcode).Msg("Instagram post fetched and decoded successfully")
	return &res, nil
}
