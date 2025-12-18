package ig

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/rs/zerolog/log"
)

func buildPostBody(shortcode string) string {
	variables := fmt.Sprintf(
		`{"shortcode":"%s","fetch_tagged_user_count":null,"hoisted_comment_id":null,"hoisted_reply_id":null}`,
		shortcode,
	)

	return url.Values{
		"fb_api_caller_class":      {"RelayModern"},
		"fb_api_req_friendly_name": {"PolarisPostActionLoadPostQueryQuery"},
		"variables":                {variables},
		"server_timestamps":        {"true"},
		"doc_id":                   {"8845758582119845"},
	}.Encode()
}

func parseInstagramVideoURL(input string) (string, bool) {
	log.Info().Str("input", input).Msg("Parsing Instagram video URL")
	re := regexp.MustCompile(`https?://www\.instagram\.com/(?:p|reel|reels?)/([A-Za-z0-9_-]+)`)
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		log.Info().Str("postID", match[1]).Msg("Instagram video URL matched")
		return match[1], true
	}
	log.Warn().Str("input", input).Msg("Instagram video URL did not match expected pattern")
	return "", false
}
