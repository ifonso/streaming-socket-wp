package spotify

import (
	"net/http"
	"time"
)

var client = &http.Client{Timeout: time.Second * 5}
