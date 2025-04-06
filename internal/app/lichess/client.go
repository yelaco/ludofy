package lichess

import "net/http"

type Client struct {
	http   http.Client
	ApiUrl string
}

func NewClient() *Client {
	return &Client{
		http:   http.Client{},
		ApiUrl: "https://lichess.org/api",
	}
}
