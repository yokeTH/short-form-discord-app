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
		"av":                       {"0"},
		"__d":                      {"www"},
		"__user":                   {"0"},
		"__a":                      {"1"},
		"__req":                    {"b"},
		"__hs":                     {"20183.HYP:instagram_web_pkg.2.1...0"},
		"dpr":                      {"3"},
		"__ccg":                    {"GOOD"},
		"__rev":                    {"1021613311"},
		"__s":                      {"hm5eih:ztapmw:x0losd"},
		"__hsi":                    {"7489787314313612244"},
		"__dyn":                    {"7xeUjG1mxu1syUbFp41twpUnwgU7SbzEdF8aUco2qwJw5ux609vCwjE1EE2Cw8G11wBz81s8hwGxu786a3a1YwBgao6C0Mo2swtUd8-U2zxe2GewGw9a361qw8Xxm16wa-0oa2-azo7u3C2u2J0bS1LwTwKG1pg2fwxyo6O1FwlA3a3zhA6bwIxe6V8aUuwm8jwhU3cyVrDyo"},
		"__csr":                    {"goMJ6MT9Z48KVkIBBvRfqKOkinBtG-FfLaRgG-lZ9Qji9XGexh7VozjHRKq5J6KVqjQdGl2pAFmvK5GWGXyk8h9GA-m6V5yF4UWagnJzazAbZ5osXuFkVeGCHG8GF4l5yp9oOezpo88PAlZ1Pxa5bxGQ7o9VrFbg-8wwxp1G2acxacGVQ00jyoE0ijonyXwfwEnwWwkA2m0dLw3tE1I80hCg8UeU4Ohox0clAhAtsM0iCA9wap4DwhS1fxW0fLhpRB51m13xC3e0h2t2H801HQw1bu02j-"},
		"__comet_req":              {"7"},
		"lsd":                      {"AVrqPT0gJDo"},
		"jazoest":                  {"2946"},
		"__spin_r":                 {"1021613311"},
		"__spin_b":                 {"trunk"},
		"__spin_t":                 {"1743852001"},
		"__crn":                    {"comet.igweb.PolarisPostRoute"},
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
