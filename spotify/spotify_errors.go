package spotify

import "fmt"

type SpotifyErrorType int

const (
	EXPIRED_TOKEN       SpotifyErrorType = 401
	BAD_OAUTH_REQUEST   SpotifyErrorType = 403
	EXCEEDED_RATE_LIMIT SpotifyErrorType = 429
)

type SpotifyError struct {
	Type SpotifyErrorType
}

func (e SpotifyError) Error() string {
	switch e.Type {
	case EXPIRED_TOKEN:
		return "Access token expired"
	case BAD_OAUTH_REQUEST:
		return "Bad OAuth request"
	case EXCEEDED_RATE_LIMIT:
		return "Exceeded rate limit"
	default:
		return fmt.Sprintf("Unknown error with code: %d", e.Type)
	}
}
