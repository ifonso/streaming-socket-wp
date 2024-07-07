package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ifonso/streaming-socket-wp/types"
)

const tokenUrl = "https://accounts.spotify.com/api/token"
const playerUrl = "https://api.spotify.com/v1/me/player/currently-playing"

type SpotifyClient struct {
	credentials struct {
		accessToken  string
		refreshToken string
		clientId     string
		clientSecret string
	}
}

func (sc *SpotifyClient) RefreshAccessToken() error {
	authCode := sc.credentials.clientId + ":" + sc.credentials.clientSecret
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(authCode))
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {sc.credentials.refreshToken},
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

	sc.credentials.accessToken = accessData.AccessToken

	return nil
}

func (sc *SpotifyClient) GetCurrentlyPlaying() (*types.SpotifyTrackResponse, error) {
	req, err := http.NewRequest(http.MethodGet, playerUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sc.credentials.accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, SpotifyError{Type: SpotifyErrorType(resp.StatusCode)}
	}

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseData := types.SpotifyTrackResponse{}

	err = json.Unmarshal(bodyData, &responseData)
	if err != nil {
		return nil, err
	}

	return &responseData, nil
}

func NewSpotifyClient(clientId, clientSecret, refreshToken string) *SpotifyClient {
	return &SpotifyClient{
		credentials: struct {
			accessToken  string
			refreshToken string
			clientId     string
			clientSecret string
		}{
			refreshToken: refreshToken,
			clientId:     clientId,
			clientSecret: clientSecret,
		},
	}
}
