package toptracks

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ifonso/streaming-socket-wp/spotify"
	"github.com/ifonso/streaming-socket-wp/types"
)

const (
	limit     = 4
	timeRange = "medium_term" // ~6 months
	cacheTTL  = 7 * 24 * time.Hour
)

// Service keeps the top tracks in memory and only asks Spotify again when
// a request arrives after the cache is older than cacheTTL.
type Service struct {
	client *spotify.SpotifyClient

	mu        sync.Mutex
	tracks    []types.SpotifyTopTrack
	fetchedAt time.Time
}

func New(client *spotify.SpotifyClient) *Service {
	return &Service{client: client}
}

// Get holds the lock while fetching so concurrent requests on an expired
// cache hit Spotify only once.
func (s *Service) Get() ([]types.SpotifyTopTrack, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.fetchedAt.IsZero() && time.Since(s.fetchedAt) < cacheTTL {
		return s.tracks, nil
	}

	items, err := s.client.GetTopTracks(limit, timeRange)
	if err != nil {
		// Serve the stale list if we have one; retry on the next request.
		if !s.fetchedAt.IsZero() {
			log.Printf("Error refreshing top tracks, serving cached: %v\n", err)
			return s.tracks, nil
		}
		return nil, err
	}

	tracks := make([]types.SpotifyTopTrack, 0, len(items))
	for i, item := range items {
		tracks = append(tracks, item.ConvertToTopTrack(i+1))
	}

	s.tracks = tracks
	s.fetchedAt = time.Now()
	return s.tracks, nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tracks, err := s.Get()
	if err != nil {
		log.Printf("Error getting top tracks: %v\n", err)
		http.Error(w, "could not get top tracks", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}
