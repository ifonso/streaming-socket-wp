package spotify

import (
	"crypto/tls"
	"net/http"
	"time"
)

var client = initHttpClient()

func initHttpClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
