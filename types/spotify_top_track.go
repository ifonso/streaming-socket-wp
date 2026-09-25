package types

type SpotifyTopTrack struct {
	Position int      `json:"position"`
	Name     string   `json:"name"`
	Artists  []string `json:"artists"`
	ImageUrl string   `json:"image_url"`
}
