package types

type SpotifyTrackResponse struct {
	Timestamp            int          `json:"timestamp"`
	ProgressMs           int          `json:"progress_ms"`
	Track                SpotifyTrack `json:"item"`
	IsPlaying            bool         `json:"is_playing"`
	CurrentlyPlayingType string       `json:"currently_playing_type"`
}

func (sr *SpotifyTrackResponse) IsTrackInPlayerType() bool {
	return sr.CurrentlyPlayingType == "track"
}
