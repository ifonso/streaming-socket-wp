package spotify

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/ifonso/streaming-socket-wp/types"
)

const tokenUrl = "https://accounts.spotify.com/api/token"
const playerUrl = "https://api.spotify.com/v1/me/player/currently-playing"
const topTracksUrl = "https://api.spotify.com/v1/me/top/tracks"

// SpotifyClient is safe for concurrent use.
type SpotifyClient struct {
	mu          sync.RWMutex
	credentials struct {
		accessToken  string
		refreshToken string
		clientId     string
		clientSecret string
	}
}

func (sc *SpotifyClient) RefreshAccessToken() error {
	sc.mu.RLock()
	authCode := sc.credentials.clientId + ":" + sc.credentials.clientSecret
	refreshToken := sc.credentials.refreshToken
	sc.mu.RUnlock()

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(authCode))
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequest(http.MethodPost, tokenUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	accessData := types.SpotifyAccessTokenResponse{}

	err = json.Unmarshal(bodyData, &accessData)
	if err != nil {
		return err
	}

	sc.mu.Lock()
	sc.credentials.accessToken = accessData.AccessToken
	// Spotify may rotate the refresh token.
	if accessData.RefreshToken != "" {
		sc.credentials.refreshToken = accessData.RefreshToken
	}
	sc.mu.Unlock()

	return nil
}

// GetCurrentlyPlaying returns nil when nothing is playing.
func (sc *SpotifyClient) GetCurrentlyPlaying() (*types.SpotifyTrackResponse, error) {
	responseData := types.SpotifyTrackResponse{}

	found, err := sc.getJSON(playerUrl, &responseData)
	if err != nil || !found {
		return nil, err
	}

	return &responseData, nil
}

func (sc *SpotifyClient) GetTopTracks(limit int, timeRange string) ([]types.SpotifyTrack, error) {
	query := url.Values{
		"limit":      {strconv.Itoa(limit)},
		"time_range": {timeRange},
	}
	responseData := types.SpotifyTopTracksResponse{}

	if _, err := sc.getJSON(topTracksUrl+"?"+query.Encode(), &responseData); err != nil {
		return nil, err
	}

	return responseData.Items, nil
}

// getJSON does an authorized GET, refreshing the access token once if it
// expired. It returns false when Spotify answers with no content.
func (sc *SpotifyClient) getJSON(url string, out any) (bool, error) {
	found, err := sc.doGetJSON(url, out)

	// TOKEN EXPIRED -> REFRESH IT AND RETRY
	if errors.Is(err, SpotifyError{Type: EXPIRED_TOKEN}) {
		if err := sc.RefreshAccessToken(); err != nil {
			return false, err
		}
		found, err = sc.doGetJSON(url, out)
	}

	return found, err
}

func (sc *SpotifyClient) doGetJSON(url string, out any) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	sc.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+sc.credentials.accessToken)
	sc.mu.RUnlock()

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, SpotifyError{Type: SpotifyErrorType(resp.StatusCode)}
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(bodyData, out); err != nil {
		return false, err
	}

	return true, nil
}

func NewSpotifyClient(clientId, clientSecret, refreshToken string) *SpotifyClient {
	sc := &SpotifyClient{}
	sc.credentials.refreshToken = refreshToken
	sc.credentials.clientId = clientId
	sc.credentials.clientSecret = clientSecret
	return sc
}
