package types

type SpotifyMusic struct {
	Id          string              `json:"id"`
	Name        string              `json:"name"`
	Images      []SpotifyAlbumImage `json:"images"`
	Artists     []string            `json:"artists"`
	Link        string              `json:"link"`
	PreviewLink string              `json:"preview_link"`
}
