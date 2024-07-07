package types

type SpotifyPlayingState struct {
	Timestamp             int          `json:"timestamp"`
	TotalTimeInSeconds    int          `json:"total_time_in_seconds"`
	ProgressTimeInSeconds int          `json:"progress_time_in_seconds"`
	IsPlaying             bool         `json:"is_playing"`
	Music                 SpotifyMusic `json:"music"`
}

func GetPlayingState(trackResponse SpotifyTrackResponse) *SpotifyPlayingState {
	return &SpotifyPlayingState{
		Timestamp:             trackResponse.Timestamp,
		TotalTimeInSeconds:    trackResponse.Track.DurationMs / 1000,
		ProgressTimeInSeconds: trackResponse.ProgressMs / 1000,
		IsPlaying:             trackResponse.IsPlaying,
		Music:                 trackResponse.Track.ConvertToMusic(),
	}
}
