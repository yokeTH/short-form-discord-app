package ig

import (
	"fmt"
	"net/url"
	"regexp"
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
	re := regexp.MustCompile(`https?://www\.instagram\.com/(?:p|reel|reels)/[A-Za-z0-9_-]+/?(?:\?[^\s]*)?`)
	match := re.FindString(input)
	if match != "" {
		return match, true
	}
	return "", false
}
