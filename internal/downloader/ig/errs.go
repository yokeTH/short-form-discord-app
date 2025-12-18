package ig

import "errors"

var (
	ErrUnsupportURL = errors.New("Unsupport url format")
	ErrNotVideo     = errors.New("Post not a video")
)
