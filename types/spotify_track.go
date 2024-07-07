package types

type SpotifyTrack struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	Album      SpotifyAlbum    `json:"album"`
	Artists    []SpotifyArtist `json:"artists"`
	DurationMs int             `json:"duration_ms"`
	Link       SpotifyLink     `json:"external_urls"`
	PreviewUrl string          `json:"preview_url"`
}

// Album
type SpotifyAlbum struct {
	Name   string              `json:"name"`
	Images []SpotifyAlbumImage `json:"images"`
}

type SpotifyAlbumImage struct {
	Url    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// Artist
type SpotifyArtist struct {
	Name   string   `json:"name"`
	Genres []string `json:"genres"`
}

// Links
type SpotifyLink struct {
	Url string `json:"spotify"`
}

func (st *SpotifyTrack) GetArtistList() []string {
	var artists []string
	for _, artist := range st.Artists {
		artists = append(artists, artist.Name)
	}
	return artists
}

func (st *SpotifyTrack) ConvertToMusic() SpotifyMusic {

	return SpotifyMusic{
		Id:          st.Id,
		Name:        st.Name,
		Images:      st.Album.Images,
		Artists:     st.GetArtistList(),
		Link:        st.Link.Url,
		PreviewLink: st.PreviewUrl,
	}
}
