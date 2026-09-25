package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ifonso/streaming-socket-wp/hub"
	"github.com/ifonso/streaming-socket-wp/poller"
	"github.com/ifonso/streaming-socket-wp/spotify"
)

const pollInterval = 10 * time.Second

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	spotifyClient := newSpotifyClient()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	h := hub.New()
	go h.Run(ctx)
	go poller.New(spotifyClient, pollInterval, h.Broadcast).Run(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.ServeWs)
	server := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("🚀🎧 Server running at port %s\n", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ListenAndServe: %v\n", err)
	}
}

func newSpotifyClient() *spotify.SpotifyClient {
	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	refreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")

	if clientId == "" || clientSecret == "" || refreshToken == "" {
		log.Fatal("Missing environment variables")
	}

	client := spotify.NewSpotifyClient(clientId, clientSecret, refreshToken)
	if err := client.RefreshAccessToken(); err != nil {
		log.Fatal(err)
	}

	return client
}
