package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ifonso/streaming-socket-wp/spotify"
	"github.com/ifonso/streaming-socket-wp/types"
)

type WsPool struct {
	clients          map[*websocket.Conn]bool
	broadcast        chan types.SpotifyPlayingState
	lastPlayingState types.SpotifyPlayingState
	spotifyClient    *spotify.SpotifyClient

	isFetching bool
}

func (wp *WsPool) startMessageBroadcast(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-wp.broadcast:
				wp.lastPlayingState = msg
				for client := range wp.clients {
					err := client.WriteJSON(msg)
					if err != nil {
						client.Close()
						delete(wp.clients, client)
					}
				}
			}
		}
	}()
}

func (wp *WsPool) startFetchingRoutine(ctx context.Context, interval time.Duration) {
	tk := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-tk.C:
				if wp.isFetching {
					continue
				}
				data := wp.getLastSpotifyState()
				if data == nil {
					data = &types.SpotifyPlayingState{
						Timestamp:             wp.lastPlayingState.Timestamp,
						TotalTimeInSeconds:    -1,
						ProgressTimeInSeconds: -1,
						IsPlaying:             false,
						Music:                 wp.lastPlayingState.Music,
					}
				}
				wp.broadcast <- *data
			case <-ctx.Done():
				tk.Stop()
				return
			}
		}
	}()
}

func (wp *WsPool) getLastSpotifyState() *types.SpotifyPlayingState {
	wp.isFetching = true
	defer func() { wp.isFetching = false }()

	trackResponse, err := wp.spotifyClient.GetCurrentlyPlaying()
	if err != nil {
		// TOKEN EXPIRED -> REFRESH IT
		if errors.Is(err, spotify.SpotifyError{Type: spotify.EXPIRED_TOKEN}) {
			err = wp.spotifyClient.RefreshAccessToken()
			if err != nil {
				log.Default().Printf("Error refreshing token: %v\n", err)
			}
			return nil
		}

		log.Default().Printf("Error getting currently playing: %v\n", err.Error())
	}

	if trackResponse == nil {
		return nil
	}

	return types.GetPlayingState(*trackResponse)
}

// Globals ---------------------------------------------------------------------

func CreateWsPool() *WsPool {
	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	refreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")

	if clientId == "" || clientSecret == "" || refreshToken == "" {
		log.Fatal("Missing environment variables")
	}

	sptClient := spotify.NewSpotifyClient(clientId, clientSecret, refreshToken)
	if err := sptClient.RefreshAccessToken(); err != nil {
		log.Fatal(err)
	}

	return &WsPool{
		clients:       make(map[*websocket.Conn]bool),
		broadcast:     make(chan types.SpotifyPlayingState),
		spotifyClient: sptClient,
	}
}

var socketPool = CreateWsPool()
var socketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Main ------------------------------------------------------------------------

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	socketPool.startMessageBroadcast(context.Background())
	socketPool.startFetchingRoutine(context.Background(), time.Second*10)

	http.HandleFunc("/ws", handleWebSocketConnections)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("ListenAndServe: %v\n", err)
	} else {
		log.Printf("🚀🎧 Server running at port %s\n", port)
	}
}

func handleWebSocketConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := socketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade: %v\n", err)
		return
	}
	defer ws.Close()

	socketPool.clients[ws] = true
	ws.WriteJSON(socketPool.lastPlayingState)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ReadMessage: %v\n", err)
			}
			delete(socketPool.clients, ws)
			break
		}
	}
}
