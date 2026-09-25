package poller

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ifonso/streaming-socket-wp/spotify"
	"github.com/ifonso/streaming-socket-wp/types"
)

// Poller periodically fetches the Spotify state and hands it to publish.
// It runs in a single goroutine, so the Spotify client and last state
// are never accessed concurrently.
type Poller struct {
	client   *spotify.SpotifyClient
	interval time.Duration
	publish  func(types.SpotifyPlayingState)
	last     types.SpotifyPlayingState
}

func New(client *spotify.SpotifyClient, interval time.Duration, publish func(types.SpotifyPlayingState)) *Poller {
	return &Poller{
		client:   client,
		interval: interval,
		publish:  publish,
	}
}

func (p *Poller) Run(ctx context.Context) {
	tk := time.NewTicker(p.interval)
	defer tk.Stop()

	for {
		p.poll()

		select {
		case <-ctx.Done():
			return
		case <-tk.C:
		}
	}
}

func (p *Poller) poll() {
	state, err := p.fetch()
	if err != nil {
		// Keep clients on the last known state instead of reporting "not playing".
		log.Printf("Error fetching Spotify state: %v\n", err)
		return
	}

	p.last = state
	p.publish(state)
}

func (p *Poller) fetch() (types.SpotifyPlayingState, error) {
	trackResponse, err := p.client.GetCurrentlyPlaying()

	// TOKEN EXPIRED -> REFRESH IT AND RETRY
	if errors.Is(err, spotify.SpotifyError{Type: spotify.EXPIRED_TOKEN}) {
		if err := p.client.RefreshAccessToken(); err != nil {
			return types.SpotifyPlayingState{}, err
		}
		trackResponse, err = p.client.GetCurrentlyPlaying()
	}

	if err != nil {
		return types.SpotifyPlayingState{}, err
	}

	if trackResponse == nil || !trackResponse.IsTrackInPlayerType() {
		return p.notPlaying(), nil
	}

	return *types.GetPlayingState(*trackResponse), nil
}

func (p *Poller) notPlaying() types.SpotifyPlayingState {
	return types.SpotifyPlayingState{
		Timestamp:             p.last.Timestamp,
		TotalTimeInSeconds:    -1,
		ProgressTimeInSeconds: -1,
		IsPlaying:             false,
		Music:                 p.last.Music,
	}
}
